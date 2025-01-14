package entity

type UploadRecord struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	OriginalFileName string `json:"originalFileName"`
	UniqueFileName   string `json:"uniqueFileName"`
	Status           string `json:"status"`
	CreatedAt        int64  `json:"createdAt"`
	ProcessedAt      int64  `json:"processedAt,omitempty"`
	OriginalFileURL  string `json:"originalFileUrl,omitempty"`
	ProcessedFileURL string `json:"processedFileUrl,omitempty"`
}
