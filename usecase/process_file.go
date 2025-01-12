package usecase

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"CutMe/domain"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type ProcessFileUseCase interface {
	domain.SQSMessageHandler
	ExtractFrames(localVideo string) (string, error)
}

type processFileUseCase struct {
	s3Client        domain.S3Client
	s3Bucket        string
	cdnDomain       string
	dynamoClient    domain.DynamoClient
	emailNotifier   domain.Notifier
	extractFramesFn func(localVideo string) (string, error) // <--- Campo funcional

}

func NewProcessFileUseCase(
	s3Client domain.S3Client,
	s3Bucket string,
	cdnDomain string,
	dynamoClient domain.DynamoClient,
	emailNotifier domain.Notifier,
	options ...func(*processFileUseCase),
) ProcessFileUseCase {
	uc := &processFileUseCase{
		s3Client:        s3Client,
		s3Bucket:        s3Bucket,
		cdnDomain:       cdnDomain,
		dynamoClient:    dynamoClient,
		emailNotifier:   emailNotifier,
		extractFramesFn: realExtractFrames,
	}

	for _, option := range options {
		option(uc)
	}

	return uc

}
func realExtractFrames(localVideo string) (string, error) {
	framesDir, err := os.MkdirTemp("", "frames_")
	if err != nil {
		return "", fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	err = ffmpeg.
		Input(localVideo).
		Filter("fps", ffmpeg.Args{"1/20"}).
		Output(filepath.Join(framesDir, "frame_%04d.jpg")).
		OverWriteOutput().
		Run()

	if err != nil {
		return "", fmt.Errorf("erro ffmpeg-go fps=1/20: %w", err)
	}
	return framesDir, nil
}

func (uc *processFileUseCase) Handle(ctx context.Context, msg domain.SQSMessage) error {
	log.Printf("Iniciando processamento do arquivo %s (ID: %s)\n", msg.FileName, msg.ID)

	originalFileURL := fmt.Sprintf("https://%s/%s", uc.cdnDomain, msg.FileName)
	uploadRecord := domain.UploadRecord{
		ID:               msg.ID,
		UserID:           msg.UserID,
		OriginalFileName: msg.FileName,
		UniqueFileName:   msg.FileName,
		Status:           "PROCESSING",
		CreatedAt:        time.Now().Unix(),
		OriginalFileURL:  originalFileURL,
	}

	if err := uc.dynamoClient.CreateOrUpdateUploadRecord(uploadRecord); err != nil {
		return fmt.Errorf("erro ao criar registro inicial no DynamoDB: %w", err)
	}

	localVideo, err := uc.s3Client.DownloadFile(uc.s3Bucket, msg.FileName)
	if err != nil {
		uc.sendFailureEmail(msg, "Erro ao baixar arquivo", err)
		return err
	}
	defer os.Remove(localVideo)

	framesDir, err := uc.extractFramesFn(localVideo)
	if err != nil {
		uc.sendFailureEmail(msg, "Erro ao extrair frames", err)
		return err
	}
	defer os.RemoveAll(framesDir)

	zipFile, err := uc.zipFrames(framesDir)
	if err != nil {
		uc.sendFailureEmail(msg, "Erro ao criar ZIP", err)
		return err
	}
	defer os.Remove(zipFile)

	zipKey := fmt.Sprintf("%s_processed.zip", msg.FileName)
	if err := uc.s3Client.UploadFile(uc.s3Bucket, zipKey, zipFile); err != nil {
		uc.sendFailureEmail(msg, "Erro ao fazer upload do ZIP", err)
		return err
	}

	uploadRecord.Status = "PROCESSED"
	uploadRecord.ProcessedAt = time.Now().Unix()
	uploadRecord.ProcessedFileURL = fmt.Sprintf("https://%s/%s", uc.cdnDomain, zipKey)

	if err := uc.dynamoClient.CreateOrUpdateUploadRecord(uploadRecord); err != nil {
		log.Printf("Erro ao atualizar registro: %v", err)
	}

	if err := uc.emailNotifier.SendSuccessEmailWithLinks(msg.UserID, msg.ID, originalFileURL, uploadRecord.ProcessedFileURL); err != nil {
		log.Printf("Erro ao enviar e-mail de sucesso: %v", err)
	}

	log.Printf("Processamento concluído para o arquivo %s\n", msg.FileName)
	return nil
}

func (uc *processFileUseCase) ExtractFrames(localVideo string) (string, error) {
	return uc.extractFrames(localVideo)
}

func (uc *processFileUseCase) extractFrames(localVideo string) (string, error) {
	framesDir, err := os.MkdirTemp("", "frames_")
	if err != nil {
		return "", fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	/**
	Aqui é basicamente como estar usando o comando ffmpeg no terminal
	para extrair frames de um vídeo. O comando equivalente seria:
		ffmpeg -i input.mp4 -vf fps=1/20 framesDir/out%d.jpg
	*/

	err = ffmpeg.
		Input(localVideo).
		// 1 frame a cada 20s
		Filter("fps", ffmpeg.Args{"1/20"}).
		Output(filepath.Join(framesDir, "frame_%04d.jpg"),
			ffmpeg.KwArgs{
				"vsync":   "vfr",
				"q:v":     2,          // qualidade
				"pix_fmt": "yuvj420p", // se necessário
			},
		).
		OverWriteOutput().
		Run()

	if err != nil {
		return "", fmt.Errorf("erro ffmpeg-go fps=1/20: %w", err)
	}

	return framesDir, nil
}

func (uc *processFileUseCase) zipFrames(framesDir string) (string, error) {
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("frames_%d.zip", time.Now().UnixNano()))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar zip file: %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(framesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(framesDir, path)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, f)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("erro ao caminhar pelos frames: %w", err)
	}

	return zipPath, nil
}

func (uc *processFileUseCase) sendFailureEmail(msg domain.SQSMessage, reason string, err error) {
	log.Printf("Erro: %s: %v\n", reason, err)
	_ = uc.emailNotifier.SendFailureEmail(msg.UserID, msg.ID, fmt.Sprintf("%s: %v", reason, err))
}
