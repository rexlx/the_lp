package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func (a *Application) AddTagHandler(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Tag added"))
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
		remoteIP := r.Header.Get("X-Real-IP")
		if remoteIP == "" {
			remoteIP = r.RemoteAddr
		}
		userAgent := r.Header.Get("User-Agent")
		tag.Memory.Lock()
		tag.Access = append(tag.Access, TagAccess{
			IP:        remoteIP,
			UserAgent: userAgent,
			Timestamp: int(time.Now().Unix()),
		})
		tag.Memory.Unlock()
		a.Memory.Lock()
		a.AccessLogs = append(a.AccessLogs, AccessLog{
			IP:        remoteIP,
			UserAgent: userAgent,
			Timestamp: int(time.Now().Unix()),
			TagID:     tag.ID,
		})
		a.Memory.Unlock()
		if err := a.DB.UpdateTag(tag); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
