package usecase

import (
	repository2 "CutMe/internal/application/repository"
	"CutMe/internal/application/service"
	entity2 "CutMe/internal/domain/entity"
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/u2takey/ffmpeg-go" // Importação correta da biblioteca
)

type processFileUseCase struct {
	s3Client      repository2.StorageClient
	s3Bucket      string
	cdnDomain     string
	dynamoClient  repository2.DBClient
	emailNotifier service.Notifier

	extractFrames func(localVideo string) (string, error)
	zipFramesFunc func(string) (string, error)
}

func NewProcessFileUseCase(
	s3Client repository2.StorageClient,
	s3Bucket string,
	cdnDomain string,
	dynamoClient repository2.DBClient,
	emailNotifier service.Notifier,
	extractFramesFunc func(string) (string, error),
	zipFramesFunc func(string) (string, error),
) ProcessFileUseCase {
	return &processFileUseCase{
		s3Client:      s3Client,
		s3Bucket:      s3Bucket,
		cdnDomain:     cdnDomain,
		dynamoClient:  dynamoClient,
		emailNotifier: emailNotifier,
		extractFrames: extractFramesFunc,
		zipFramesFunc: zipFramesFunc,
	}
}

func (uc *processFileUseCase) Handle(ctx context.Context, msg entity2.Message) error {
	log.Printf("Iniciando processamento do arquivo %s (ID: %s)\n", msg.FileName, msg.ID)

	originalFileURL := fmt.Sprintf("https://%s/%s", uc.cdnDomain, msg.FileName)
	uploadRecord := entity2.UploadRecord{
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

	framesDir, err := uc.extractFrames(localVideo)
	if err != nil {
		uc.sendFailureEmail(msg, "Erro ao extrair frames", err)
		return err
	}
	defer os.RemoveAll(framesDir)

	var zipFile string

	if uc.zipFramesFunc != nil {
		zipFile, err = uc.zipFramesFunc(framesDir)
	} else {
		zipFile, err = uc.zipFrames(framesDir) // chama método real
	}
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
	framesDir, err := os.MkdirTemp("", "frames_")
	if err != nil {
		return "", fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	err = ffmpeg_go.
		Input(localVideo).
		Filter("fps", ffmpeg_go.Args{"1/20"}). // 1 frame a cada 20 segundos
		Output(filepath.Join(framesDir, "frame_%04d.jpg"),
			ffmpeg_go.KwArgs{
				"vsync":   "vfr",
				"q:v":     2,
				"pix_fmt": "yuvj420p",
			},
		).
		OverWriteOutput().
		Run()

	if err != nil {
		return "", fmt.Errorf("erro ao processar frames com ffmpeg-go: %w", err)
	}

	return framesDir, nil
}

func (uc *processFileUseCase) zipFrames(framesDir string) (string, error) {
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("frames_%d.zip", time.Now().UnixNano()))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo ZIP: %w", err)
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

func (uc *processFileUseCase) sendFailureEmail(msg entity2.Message, reason string, err error) {
	log.Printf("Erro: %s: %v\n", reason, err)
	_ = uc.emailNotifier.SendFailureEmail(msg.UserID, msg.ID, fmt.Sprintf("%s: %v", reason, err))
}

func DefaultExtractFrames(localVideo string) (string, error) {
	framesDir, err := os.MkdirTemp("", "frames_")
	if err != nil {
		return "", fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	err = ffmpeg_go.
		Input(localVideo).
		Filter("fps", ffmpeg_go.Args{"1/20"}).
		Output(filepath.Join(framesDir, "frame_%04d.jpg"),
			ffmpeg_go.KwArgs{
				"vsync":   "vfr",
				"q:v":     2,
				"pix_fmt": "yuvj420p",
			},
		).
		OverWriteOutput().
		Run()

	if err != nil {
		return "", fmt.Errorf("erro ao processar frames com ffmpeg-go: %w", err)
	}

	return framesDir, nil
}

func DefaultZipFrames(framesDir string) (string, error) {
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("frames_%d.zip", time.Now().UnixNano()))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo ZIP: %w", err)
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
