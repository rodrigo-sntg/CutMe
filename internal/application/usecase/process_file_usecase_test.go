package usecase

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"CutMe/internal/application/mocks"
	"CutMe/internal/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Funções fake para simular extração de frames e zipagem
func fakeExtractFramesSuccess(localVideo string) (string, error) {
	dir := filepath.Join(os.TempDir(), "fake_frames_success")
	_ = os.MkdirAll(dir, 0755)
	f, _ := os.Create(filepath.Join(dir, "frame_0001.jpg"))
	_ = f.Close()
	return dir, nil
}

func fakeExtractFramesFail(localVideo string) (string, error) {
	return "", errors.New("erro ao extrair frames (fake)")
}

func fakeZipFramesSuccess(framesDir string) (string, error) {
	zipPath := filepath.Join(os.TempDir(), "fake_frames.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return zipPath, nil
}

func fakeZipFramesFail(framesDir string) (string, error) {
	return "", errors.New("erro ao criar ZIP (fake)")
}

func TestProcessFileUseCase_Handle_Success(t *testing.T) {
	storageMock := &mocks.StorageClientMock{}
	dbMock := &mocks.DBClientMock{}
	notifierMock := &mocks.NotifierMock{}

	storageMock.On("DownloadFile", "myBucket", "video.mp4").Return("/tmp/fake_video.mp4", nil)
	storageMock.On("UploadFile", "myBucket", "video.mp4_processed.zip", mock.AnythingOfType("string")).Return(nil)
	dbMock.On("CreateOrUpdateUploadRecord", mock.Anything).Return(nil).Twice()
	notifierMock.On("SendSuccessEmailWithLinks", "user123", "id123", "https://cdnDomain/video.mp4", mock.AnythingOfType("string")).Return(nil)

	uc := NewProcessFileUseCase(
		storageMock, "myBucket", "cdnDomain", dbMock, notifierMock,
		fakeExtractFramesSuccess, fakeZipFramesSuccess,
	)

	msg := entity.Message{
		ID:       "id123",
		UserID:   "user123",
		FileName: "video.mp4",
	}

	err := uc.Handle(context.Background(), msg)
	assert.NoError(t, err)
	storageMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	notifierMock.AssertExpectations(t)
}

func TestProcessFileUseCase_Handle_FailZipFrames(t *testing.T) {
	storageMock := &mocks.StorageClientMock{}
	dbMock := &mocks.DBClientMock{}
	notifierMock := &mocks.NotifierMock{}

	dbMock.On("CreateOrUpdateUploadRecord", mock.Anything).Return(nil).Once()
	storageMock.On("DownloadFile", "myBucket", "video.mp4").Return("/tmp/fake_video.mp4", nil)
	notifierMock.On("SendFailureEmail", "user123", "id123", mock.AnythingOfType("string")).Return(nil)

	uc := NewProcessFileUseCase(
		storageMock, "myBucket", "cdnDomain", dbMock, notifierMock,
		fakeExtractFramesSuccess, fakeZipFramesFail,
	)

	msg := entity.Message{
		ID:       "id123",
		UserID:   "user123",
		FileName: "video.mp4",
	}

	err := uc.Handle(context.Background(), msg)
	assert.ErrorContains(t, err, "erro ao criar ZIP (fake)")
	storageMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	notifierMock.AssertExpectations(t)
}

func TestProcessFileUseCase_Handle_FailMissingFile(t *testing.T) {
	storageMock := &mocks.StorageClientMock{}
	dbMock := &mocks.DBClientMock{}
	notifierMock := &mocks.NotifierMock{}

	// Registro inicial no DB
	dbMock.
		On("CreateOrUpdateUploadRecord", mock.Anything).
		Return(nil).
		Once()

	// Simula falha no download porque o arquivo não existe
	storageMock.
		On("DownloadFile", "myBucket", "missing_file.mp4").
		Return("", errors.New("arquivo não encontrado"))

	// Deve enviar e-mail de falha
	notifierMock.
		On("SendFailureEmail", "user123", "id123", mock.AnythingOfType("string")).
		Return(nil)

	uc := NewProcessFileUseCase(
		storageMock,
		"myBucket",
		"cdnDomain",
		dbMock,
		notifierMock,
		fakeExtractFramesSuccess,
		fakeZipFramesSuccess,
	)

	msg := entity.Message{
		ID:       "id123",
		UserID:   "user123",
		FileName: "missing_file.mp4",
	}

	err := uc.Handle(context.Background(), msg)
	assert.ErrorContains(t, err, "arquivo não encontrado")
	dbMock.AssertExpectations(t)
	storageMock.AssertExpectations(t)
	notifierMock.AssertExpectations(t)
}

func TestProcessFileUseCase_Handle_FailEmptyFrames(t *testing.T) {
	storageMock := &mocks.StorageClientMock{}
	dbMock := &mocks.DBClientMock{}
	notifierMock := &mocks.NotifierMock{}

	// Registro inicial no DB
	dbMock.
		On("CreateOrUpdateUploadRecord", mock.Anything).
		Return(nil).
		Once()

	// Download do arquivo ok
	storageMock.
		On("DownloadFile", "myBucket", "video.mp4").
		Return("/tmp/fake_video.mp4", nil)

	// Função fake retorna diretório vazio para frames
	extractFramesEmpty := func(localVideo string) (string, error) {
		dir := filepath.Join(os.TempDir(), "empty_frames")
		_ = os.MkdirAll(dir, 0755)
		return dir, nil
	}

	// Deve enviar e-mail de falha ao criar ZIP
	notifierMock.
		On("SendFailureEmail", "user123", "id123", mock.AnythingOfType("string")).
		Return(nil)

	uc := NewProcessFileUseCase(
		storageMock,
		"myBucket",
		"cdnDomain",
		dbMock,
		notifierMock,
		extractFramesEmpty,
		fakeZipFramesFail,
	)

	msg := entity.Message{
		ID:       "id123",
		UserID:   "user123",
		FileName: "video.mp4",
	}

	err := uc.Handle(context.Background(), msg)
	assert.ErrorContains(t, err, "erro ao criar ZIP (fake)")
	dbMock.AssertExpectations(t)
	storageMock.AssertExpectations(t)
	notifierMock.AssertExpectations(t)
}

func TestProcessFileUseCase_Handle_FailDatabaseUnavailable(t *testing.T) {
	storageMock := &mocks.StorageClientMock{}
	dbMock := &mocks.DBClientMock{}
	notifierMock := &mocks.NotifierMock{}

	// Simula falha ao registrar inicial no banco de dados
	dbMock.
		On("CreateOrUpdateUploadRecord", mock.Anything).
		Return(errors.New("banco de dados indisponível"))

	uc := NewProcessFileUseCase(
		storageMock,
		"myBucket",
		"cdnDomain",
		dbMock,
		notifierMock,
		fakeExtractFramesSuccess,
		fakeZipFramesSuccess,
	)

	msg := entity.Message{
		ID:       "id123",
		UserID:   "user123",
		FileName: "video.mp4",
	}

	err := uc.Handle(context.Background(), msg)
	assert.ErrorContains(t, err, "banco de dados indisponível")
	dbMock.AssertExpectations(t)
	storageMock.AssertNotCalled(t, "DownloadFile", mock.Anything, mock.Anything)
	notifierMock.AssertNotCalled(t, "SendFailureEmail", mock.Anything, mock.Anything, mock.Anything)
}

func TestProcessFileUseCase_Handle_FailAllSteps(t *testing.T) {
	storageMock := &mocks.StorageClientMock{}
	dbMock := &mocks.DBClientMock{}
	notifierMock := &mocks.NotifierMock{}

	// Registro inicial no DB falha
	dbMock.
		On("CreateOrUpdateUploadRecord", mock.Anything).
		Return(errors.New("falha ao registrar no banco"))

	// Download falha
	storageMock.
		On("DownloadFile", "myBucket", "video.mp4").
		Return("", errors.New("erro ao baixar arquivo"))

	// Deve enviar e-mail de falha
	notifierMock.
		On("SendFailureEmail", "user123", "id123", mock.AnythingOfType("string")).
		Return(nil)

	uc := NewProcessFileUseCase(
		storageMock,
		"myBucket",
		"cdnDomain",
		dbMock,
		notifierMock,
		fakeExtractFramesFail,
		fakeZipFramesFail,
	)

	msg := entity.Message{
		ID:       "id123",
		UserID:   "user123",
		FileName: "video.mp4",
	}

	err := uc.Handle(context.Background(), msg)
	assert.ErrorContains(t, err, "falha ao registrar no banco")
	assert.ErrorContains(t, err, "erro ao criar registro inicial no DynamoDB: falha ao registrar no banco")

	// Verificar se as expectativas foram cumpridas
	dbMock.AssertExpectations(t)
	storageMock.AssertNotCalled(t, "DownloadFile", mock.Anything, mock.Anything)

}

func setupFakeVideoFile(path string) error {
	// Usa o ffmpeg para criar um vídeo válido
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "color=c=blue:s=320x240:d=1", "-y", path)
	return cmd.Run()
}

func cleanupFakeFile(path string) {
	os.Remove(path)
}

func TestProcessFileUseCase_ExtractFrames_Success(t *testing.T) {
	fakeVideo := "/tmp/fake_video.mp4"
	err := setupFakeVideoFile(fakeVideo)
	assert.NoError(t, err)
	defer cleanupFakeFile(fakeVideo)

	uc := &processFileUseCase{}
	dir, err := uc.ExtractFrames(fakeVideo)
	assert.NoError(t, err)
	assert.DirExists(t, dir)
	os.RemoveAll(dir)
}

func TestProcessFileUseCase_ExtractFrames_Fail(t *testing.T) {
	uc := &processFileUseCase{}
	_, err := uc.ExtractFrames("/invalid/path/video.mp4")
	assert.ErrorContains(t, err, "erro ao processar frames com ffmpeg-go")
}

func TestProcessFileUseCase_ZipFrames_Success(t *testing.T) {
	framesDir := filepath.Join(os.TempDir(), "test_frames")
	_ = os.MkdirAll(framesDir, 0755)
	_ = os.WriteFile(filepath.Join(framesDir, "frame_0001.jpg"), []byte("fake frame data"), 0644)
	defer os.RemoveAll(framesDir)

	uc := &processFileUseCase{}
	zipFile, err := uc.zipFrames(framesDir)
	assert.NoError(t, err)
	assert.FileExists(t, zipFile)
	os.Remove(zipFile) // Limpeza
}

func TestProcessFileUseCase_ZipFrames_Fail(t *testing.T) {
	uc := &processFileUseCase{}
	_, err := uc.zipFrames("/invalid/path")
	assert.ErrorContains(t, err, "erro ao caminhar pelos frames")
}

func TestDefaultExtractFrames(t *testing.T) {
	fakeVideo := "/tmp/fake_video.mp4"

	// Configura um vídeo válido para o teste
	err := setupFakeVideoFile(fakeVideo)
	assert.NoError(t, err, "Failed to create fake video file")
	defer cleanupFakeFile(fakeVideo)

	framesDir, err := DefaultExtractFrames(fakeVideo)
	assert.NoError(t, err, "DefaultExtractFrames failed")
	assert.DirExists(t, framesDir, "Frames directory does not exist")
	os.RemoveAll(framesDir)
}
func TestDefaultZipFrames(t *testing.T) {
	framesDir := filepath.Join(os.TempDir(), "test_frames")
	_ = os.MkdirAll(framesDir, 0755)
	_ = os.WriteFile(filepath.Join(framesDir, "frame_0001.jpg"), []byte("fake frame data"), 0644)
	defer os.RemoveAll(framesDir)

	zipFile, err := DefaultZipFrames(framesDir)
	assert.NoError(t, err)
	assert.FileExists(t, zipFile)
	os.Remove(zipFile)
}
