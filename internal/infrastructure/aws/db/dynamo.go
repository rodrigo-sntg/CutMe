package db

import (
	"CutMe/internal/application/repository"
	"CutMe/internal/domain/entity"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dynamoClient struct {
	Svc       dynamodbiface.DynamoDBAPI
	TableName string
}

func NewDynamoClient(sess *session.Session, tableName string) repository.DBClient {
	return &dynamoClient{
		Svc:       dynamodb.New(sess),
		TableName: tableName,
	}
}

func (d *dynamoClient) UpdateStatus(id, status string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("SET #st = :val"),
		ExpressionAttributeNames: map[string]*string{
			"#st": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": {
				S: aws.String(status),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := d.Svc.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status no Dynamo: %w", err)
	}
	return nil
}

func (d *dynamoClient) UpdateUploadRecord(record entity.UploadRecord) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(record.ID),
			},
		},
		UpdateExpression: aws.String("SET UniqueFileName = :unique, Status = :status, ProcessedAt = :processed"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":unique": {
				S: aws.String(record.UniqueFileName),
			},
			":status": {
				S: aws.String(record.Status),
			},
			":processed": {
				N: aws.String(fmt.Sprintf("%d", record.ProcessedAt)),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := d.Svc.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("erro ao atualizar registro no DynamoDB: %w", err)
	}
	return nil
}

func (d *dynamoClient) CreateUploadRecord(record entity.UploadRecord) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.TableName),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(record.ID),
			},
			"userId": {
				S: aws.String(record.UserID),
			},
			"originalFileName": {
				S: aws.String(record.OriginalFileName),
			},
			"uniqueFileName": {
				S: aws.String(record.UniqueFileName),
			},
			"status": {
				S: aws.String(record.Status),
			},
			"createdAt": {
				N: aws.String(fmt.Sprintf("%d", record.CreatedAt)),
			},
			"originalFileURL": {
				S: aws.String(record.OriginalFileURL),
			},
			"processedFileURL": {
				S: aws.String(record.ProcessedFileURL),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	}

	_, err := d.Svc.PutItem(input)

	if err != nil {
		if isConditionalCheckFailed(err) {
			return d.UpdateStatus(record.ID, "PROCESSING")
		}
		return fmt.Errorf("erro ao criar registro no DynamoDB: %w", err)
	}

	return nil
}

func isConditionalCheckFailed(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException
	}
	// Verificação adicional para erros simulados
	return err != nil && err.Error() == dynamodb.ErrCodeConditionalCheckFailedException
}

func (d *dynamoClient) CreateOrUpdateUploadRecord(record entity.UploadRecord) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(record.ID),
			},
		},
		UpdateExpression: aws.String(`
			SET 
				originalFileName = :originalFileName, 
				uniqueFileName = :uniqueFileName, 
				userId = :userId,
				#st = :status, 
        		originalFileURL     = :originalFileURL,
		        processedFileURL    = :processedFileURL,
				createdAt = if_not_exists(createdAt, :createdAt),
				processedAt = :processedAt
		`),
		ExpressionAttributeNames: map[string]*string{
			"#st": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":originalFileName": {
				S: aws.String(record.OriginalFileName),
			},
			":userId": {
				S: aws.String(record.UserID),
			},
			":uniqueFileName": {
				S: aws.String(record.UniqueFileName),
			},
			":status": {
				S: aws.String(record.Status),
			},
			":originalFileURL": {
				S: aws.String(record.OriginalFileURL),
			},
			":processedFileURL": {
				S: aws.String(record.ProcessedFileURL),
			},
			":createdAt": {
				N: aws.String(fmt.Sprintf("%d", record.CreatedAt)),
			},
			":processedAt": {
				N: aws.String(fmt.Sprintf("%d", record.ProcessedAt)),
			},
		},
		ReturnValues: aws.String("ALL_NEW"),
	}

	_, err := d.Svc.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("erro ao criar ou atualizar registro no DynamoDB: %w", err)
	}

	return nil

}

func (d *dynamoClient) GetUploads(status string) ([]entity.UploadRecord, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(d.TableName),
	}

	if status != "" {
		input.FilterExpression = aws.String("#st = :status")
		input.ExpressionAttributeNames = map[string]*string{
			"#st": aws.String("status"),
		}
		input.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(status),
			},
		}
	}

	result, err := d.Svc.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear registros no DynamoDB: %w", err)
	}

	var uploads []entity.UploadRecord
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &uploads); err != nil {
		return nil, fmt.Errorf("erro ao desserializar registros: %w", err)
	}

	return uploads, nil
}

func (d *dynamoClient) GetUploadByID(id string) (*entity.UploadRecord, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	result, err := d.Svc.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar registro no DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("registro com ID %s não encontrado", id)
	}

	var upload entity.UploadRecord
	if err := dynamodbattribute.UnmarshalMap(result.Item, &upload); err != nil {
		return nil, fmt.Errorf("erro ao desserializar registro: %w", err)
	}

	return &upload, nil
}

func (d *dynamoClient) GetUploadsByUserID(userID string, status string) ([]entity.UploadRecord, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.TableName),
		IndexName:              aws.String("UserID-index"),
		KeyConditionExpression: aws.String("userId = :userId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userId": {
				S: aws.String(userID),
			},
		},
	}

	if status != "" {
		input.FilterExpression = aws.String("#st = :status")
		input.ExpressionAttributeNames = map[string]*string{
			"#st": aws.String("status"),
		}
		input.ExpressionAttributeValues[":status"] = &dynamodb.AttributeValue{
			S: aws.String(status),
		}
	}

	result, err := d.Svc.Query(input)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar registros no DynamoDB: %w", err)
	}

	var uploads []entity.UploadRecord
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &uploads); err != nil {
		return nil, fmt.Errorf("erro ao desserializar registros: %w", err)
	}

	return uploads, nil
}
