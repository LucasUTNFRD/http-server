package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// Check r.TransferEnconding()
func echoHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/echo/") {
		http.NotFound(w, r)
		return
	}
	message := strings.TrimPrefix(r.URL.Path, "/echo/")

	//TODO add support for the Accept-Encoding and Content-Encoding headers.
	//TODO add support for gzip compression
	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "gzip") {
		// Create a buffer to hold the compressed data
		var buf strings.Builder
		gzipBufWriter := gzip.NewWriter(&buf)

		// Write the message to the gzip buffer
		_, err := gzipBufWriter.Write([]byte(message))
		if err != nil {
			http.Error(w, "Failed to compress data", http.StatusInternalServerError)
			return
		}
		gzipBufWriter.Close()

		// Set headers before writing to the response writer
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.WriteHeader(http.StatusOK)

		// Write the compressed data to the response writer
		_, err = w.Write([]byte(buf.String()))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		return
	}
	//default response
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(message)))
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, message)
}

func userAgentHandler(w http.ResponseWriter, r *http.Request) {
	userAgent := r.UserAgent()
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(userAgent)))
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, userAgent)
}

func fileHandler(w http.ResponseWriter, r *http.Request, dir string) {
	log.Println("debugg")
	fileName := strings.TrimPrefix(r.URL.Path, "/files/")
	filePath := filepath.Join(dir, fileName)

	switch r.Method {
	case http.MethodGet:
		getFile(w, r, filePath)
	case http.MethodPost:
		postFile(w, r, filePath)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

// TODO: create the file put the data inside and return 201
func postFile(w http.ResponseWriter, r *http.Request, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Ensure the file is closed when the function returns
	defer file.Close()

	// Copy the request body into the newly created file
	_, err = io.Copy(file, r.Body)
	if err != nil {
		// If there's an error copying the data, return a 500 Internal Server Error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("aca deberia retornal el status 201")
	w.WriteHeader(http.StatusCreated)

	fmt.Fprint(w)

}
func getFile(w http.ResponseWriter, r *http.Request, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		log.Println("not found")
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, filePath)
}
