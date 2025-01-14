package service

type Notifier interface {
	SendSuccessEmailWithLinks(to, uploadID, originalFileURL, processedFileURL string) error
	SendFailureEmail(to, uploadID, errorMsg string) error
}
