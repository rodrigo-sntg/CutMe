package signed_url

import (
	"CutMe/internal/application/mocks"
	"CutMe/internal/domain/entity"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"time"

	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateUploadURLHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGenerator := mocks.NewMockPresignedURLGenerator(ctrl)
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockGenerator.EXPECT().
			GeneratePreSignedURL("test-file.txt", "text/plain", "123").
			Return("http://signed-url.com", "unique-id-123", nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
		})
		router.POST("/generate-url", GenerateUploadURLHandler(mockGenerator))

		body := map[string]string{
			"fileName": "test-file.txt",
			"fileType": "text/plain",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/generate-url", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "http://signed-url.com", response["signedUrl"])
		assert.Equal(t, "unique-id-123", response["uniqueId"])
	})

	t.Run("missing userID", func(t *testing.T) {
		router := gin.New()
		router.POST("/generate-url", GenerateUploadURLHandler(mockGenerator))

		body := map[string]string{
			"fileName": "test-file.txt",
			"fileType": "text/plain",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/generate-url", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("invalid parameters", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
		})
		router.POST("/generate-url", GenerateUploadURLHandler(mockGenerator))

		body := map[string]string{
			"fileType": "text/plain", // Missing fileName
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/generate-url", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("error generating signed URL", func(t *testing.T) {
		mockGenerator.EXPECT().
			GeneratePreSignedURL("test-file.txt", "text/plain", "123").
			Return("", "", errors.New("failed to generate URL"))

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
		})
		router.POST("/generate-url", GenerateUploadURLHandler(mockGenerator))

		body := map[string]string{
			"fileName": "test-file.txt",
			"fileType": "text/plain",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/generate-url", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestNewS3SignedURLGenerator(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		t.Fatalf("failed to create AWS session: %v", err)
	}

	generator := NewS3SignedURLGenerator(sess, "test-bucket", 15*time.Minute)
	assert.NotNil(t, generator)
	assert.Equal(t, "test-bucket", generator.bucketName)
	assert.Equal(t, 15*time.Minute, generator.urlValidity)
}

func TestGetBucketName(t *testing.T) {
	generator := &SignedURLGenerator{bucketName: "test-bucket"}
	assert.Equal(t, "test-bucket", generator.GetBucketName())
}

func TestGetURLValidity(t *testing.T) {
	generator := &SignedURLGenerator{urlValidity: 15 * time.Minute}
	assert.Equal(t, 15*time.Minute, generator.GetURLValidity())
}

func TestUploadsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDBClient(ctrl)
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockDynamo.EXPECT().
			GetUploadsByUserID("123", "PROCESSING").
			Return([]entity.UploadRecord{
				{ID: "upload-1", UserID: "123", Status: "PROCESSING"},
			}, nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
		})
		router.GET("/uploads", UploadsHandler(mockDynamo))

		req := httptest.NewRequest(http.MethodGet, "/uploads?status=PROCESSING", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		router := gin.New()
		router.GET("/uploads", UploadsHandler(mockDynamo))

		req := httptest.NewRequest(http.MethodGet, "/uploads", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("error fetching uploads", func(t *testing.T) {
		mockDynamo.EXPECT().
			GetUploadsByUserID("123", "PROCESSING").
			Return(nil, errors.New("DB error"))

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "123")
		})
		router.GET("/uploads", UploadsHandler(mockDynamo))

		req := httptest.NewRequest(http.MethodGet, "/uploads?status=PROCESSING", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestCreateUploadHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDBClient(ctrl)
	mockS3 := mocks.NewMockStorageClient(ctrl)
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		router := gin.New()
		router.POST("/create-upload", CreateUploadHandler(mockDynamo, mockS3))

		body := map[string]string{
			"fileName": "test-file.txt",
			"fileType": "text/plain",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/create-upload", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)
	})

	t.Run("invalid parameters", func(t *testing.T) {
		router := gin.New()
		router.POST("/create-upload", CreateUploadHandler(mockDynamo, mockS3))

		body := map[string]string{
			"fileType": "text/plain",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/create-upload", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
