package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	assetBucket := os.Getenv("INFRAI_BUCKET")
	if assetBucket == "" {
		log.Fatal("INFRAI_BUCKET is required")
	}
	client := newInfraiClient(apiKey)

	service := uploadService{bucket: assetBucket, signer: client}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload-url", uploadHandler(service))
	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("logistics upload service listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func uploadHandler(service uploadService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in uploadRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		grant, err := service.grant(r.Context(), in)
		if err != nil {
			status := http.StatusBadRequest
			var remote *InfraiError
			if errors.As(err, &remote) && remote.Status >= 400 && remote.Status < 500 {
				status = remote.Status
			}
			if errors.As(err, &remote) && remote.Status >= 500 {
				status = http.StatusBadGateway
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, grant)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
