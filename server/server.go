package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/starlinglab/archive-backend/db"
	"github.com/starlinglab/archive-backend/types"
)

type Server struct {
	mux http.ServeMux
}

func NewServer() *Server {
	s := &Server{}
	s.mux.HandleFunc("/v1/store", store)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func store(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096) // Max 4 KiB of JSON

	var sr types.StorageRequest
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &sr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: calculate providers

	if err := db.AddToQueue(sr.Hash, []string{}, &sr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
