package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

		err := a.WriteToDisk(fmt.Sprintf("./static/%s", filename), fileData.Bytes())
		if err != nil {
			fmt.Println("Error writing to disk:", err)
			return // Or handle error appropriately
		}

		modifiedFilename := filepath.Base(filename)
		modifiedFilenameWithoutExt := modifiedFilename[:len(modifiedFilename)-len(filepath.Ext(modifiedFilename))]
		modifiedFilename = modifiedFilenameWithoutExt + "_new.pdf"

		err = RunBashScript("./scripts/call_add_py.sh", fmt.Sprintf("./static/%s", filename), uid)
		if err != nil {
			fmt.Println("Error running script:", err)
			return // Or handle error appropriately
		}

		modifiedFilePath := fmt.Sprintf("./static/%s", modifiedFilename)
		modifiedFile, err := os.Open(modifiedFilePath)
		if err != nil {
			fmt.Println("Error opening modified file:", err)
			return // Or handle error appropriately
		}
		defer modifiedFile.Close()

		hash, err := CalculateSHA256(modifiedFile)
		if err != nil {
			fmt.Println("Error calculating SHA256:", err)
			return // Or handle error appropriately
		}

		tag := &Tag{
			Hash:    hash,
			ID:      uid,
			URL:     fmt.Sprintf("%s/%s", a.FQDN, uid),
			History: []TagHistoryItem{},
			Access:  []TagAccess{},
		}
		a.AddTag(tag)

		fmt.Println("Removed files:", filename, modifiedFilename)
		fmt.Println("File written successfully:", filename)
	}
	out, err := json.Marshal(UploadResponse)
	if err != nil {
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}

func CalculateSHA256(file *os.File) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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
