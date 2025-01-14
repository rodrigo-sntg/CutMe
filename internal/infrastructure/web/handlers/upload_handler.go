package handlers

import (
	"CutMe/internal/application/repository"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type UploadHandler struct {
	DynamoClient repository.DBClient
}

func NewUploadHandler(dynamoClient repository.DBClient) *UploadHandler {
	return &UploadHandler{DynamoClient: dynamoClient}
}

func (h *UploadHandler) ListUploads(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	uploads, err := h.DynamoClient.GetUploads(status)
	if err != nil {
		log.Printf("Erro ao listar uploads: %v", err)
		http.Error(w, "Erro ao buscar uploads", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uploads)
}

func (h *UploadHandler) GetUpload(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	upload, err := h.DynamoClient.GetUploadByID(id)
	if err != nil {
		log.Printf("Erro ao buscar upload: %v", err)
		http.Error(w, "Registro não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upload)
}
