package main

import (
	"log"
	"sync"
)

type Tag struct {
	Username string           `json:"username"`
	FilePath string           `json:"file_path"`
	ID       string           `json:"id"`
	ClientID string           `json:"client_id"`
	Hash     string           `json:"hash"`
	URL      string           `json:"url"`
	Created  int              `json:"created"`
	History  []TagHistoryItem `json:"history"`
	Access   []TagAccess      `json:"access"`
	Memory   *sync.RWMutex    `json:"-"`
}

type TagAccess struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Timestamp int    `json:"timestamp"` // Unix timestamp
}

type TagHistoryItem struct {
	ClientID string `json:"client_id"`
	Hash     string `json:"hash"`
	Created  int    `json:"created"`
}

func NewTag(id, clientID, hash string, created int) *Tag {
	return &Tag{
		Memory:   &sync.RWMutex{},
		ID:       id,
		ClientID: clientID,
		Hash:     hash,
		Created:  created,
		History:  []TagHistoryItem{},
		Access:   []TagAccess{},
	}
}

func (t *Tag) AddHistory(clientID, hash string, created int) {
	// t.Memory.Lock()
	// defer t.Memory.Unlock()
	if len(t.History) > 149 {
		log.Printf("tag %s history is full, removing oldest item", t.ID)
		t.History = t.History[1:]
	}
	t.History = append(t.History, TagHistoryItem{
		ClientID: clientID,
		Hash:     hash,
		Created:  created,
	})
}

func (t *Tag) AddAccess(ip, userAgent string, timestamp int) {
	// t.Memory.Lock()
	// defer t.Memory.Unlock()
	if len(t.Access) > 149 {
		log.Printf("tag %s access is full, removing oldest item", t.ID)
		t.Access = t.Access[1:]
	}
	t.Access = append(t.Access, TagAccess{
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: timestamp,
	})
}

func (t *Tag) GetHistory() []TagHistoryItem {
	t.Memory.RLock()
	defer t.Memory.RUnlock()
	return t.History
}
