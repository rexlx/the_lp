package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (a *Application) AddTagHandler(w http.ResponseWriter, r *http.Request) {
	out := `<w:instrText xml:space="preserve"> INCLUDEPICTURE \d "%v" \* MERGEFORMATINET </w:instrText>`
	tag := &Tag{}
	if err := json.NewDecoder(r.Body).Decode(tag); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if tag.ID == "" || tag.Hash == "" {
		http.Error(w, "ID and hash is required", http.StatusBadRequest)
		return
	}
	if tag.Created == 0 {
		tag.Created = int(time.Now().Unix())
	}
	a.AddTag(tag)
	if tag.URL == "" {
		tag.URL = fmt.Sprintf("%s/%s", a.FQDN, tag.ID)
	}
	w.WriteHeader(http.StatusCreated)
	// res := make(map[string]string)
	wordString := fmt.Sprintf(out, tag.URL)
	type Response struct {
		Data string `json:"data"`
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Data: wordString})
}

type TagQuery struct {
	ID string `json:"id"`
}

func (a *Application) TagExistsHandler(w http.ResponseWriter, r *http.Request) {
	query := &TagQuery{}
	if err := json.NewDecoder(r.Body).Decode(query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tag := a.GetTag(query.ID)
	if tag == nil {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tag exists"))
}

func (a *Application) GetTagHandler(w http.ResponseWriter, r *http.Request) {
	query := &TagQuery{}
	if err := json.NewDecoder(r.Body).Decode(query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tag := a.GetTag(query.ID)
	if tag == nil {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (a *Application) tagHandler(tag *Tag) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		remoteIP := r.Header.Get("X-Forwarded-For")
		if remoteIP == "" {
			remoteIP = r.RemoteAddr
		}
		userAgent := r.Header.Get("User-Agent")
		a.Logger.Info("Tag accessed", zap.String("tag_id", tag.ID), zap.String("remote_ip", remoteIP), zap.String("user_agent", userAgent))
		go tag.AddAccess(remoteIP, userAgent, int(time.Now().Unix()))
		a.AddAccess(&AccessLog{
			IP:        remoteIP,
			UserAgent: userAgent,
			Timestamp: int(time.Now().Unix()),
			TagID:     tag.ID,
		})
		if err := a.DB.UpdateTag(tag); err != nil {
			// log.Println("error updating tag", err)
			a.Logger.Error("error updating tag", zap.String("tag_id", tag.ID), zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodPost {
			// Handle form submission
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("OK")) // Or a simple response indicating success.
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tag)
	}
}

func (a *Application) AccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	a.Memory.RLock()
	defer a.Memory.RUnlock()
	json.NewEncoder(w).Encode(a.AccessLogs)
}
