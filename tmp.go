package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type uploadResponse struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

func (a *Application) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	var complete bool
	var fileData bytes.Buffer
	var UploadResponse uploadResponse
	UploadResponse.Status = "incomplete"
	// Copy the request body (file data) to the buffer
	_, err := io.Copy(&fileData, r.Body)
	if err != nil {
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}
	chunkSize := r.ContentLength
	filename := r.Header.Get("X-filename")
	filename = filepath.Base(filename)

	lastChunk := r.Header.Get("X-last-chunk")
	fmt.Println(chunkSize, filename, lastChunk)

	if lastChunk == "true" {
		fmt.Println("last chunk", chunkSize, filename)
		complete = true
	}

	if complete {
		uid := uuid.New().String()
		UploadResponse.ID = uid
		UploadResponse.Status = "complete"
		go a.WriteToDisk(fmt.Sprintf("./static/%s", filename), fileData.Bytes())

	}
	out, err := json.Marshal(UploadResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (a *Application) WriteToDisk(filename string, data []byte) error {
	// Create the file on disk
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	fmt.Println("File written successfully:", filename)
	return nil
}
