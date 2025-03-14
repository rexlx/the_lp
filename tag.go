package main

import "sync"

type Tag struct {
	Memory   *sync.RWMutex    `json:"-"`
	ID       string           `json:"id"`
	ClientID string           `json:"client_id"`
	Hash     string           `json:"hash"`
	URL      string           `json:"url"`
	Created  int              `json:"created"`
	History  []TagHistoryItem `json:"history"`
	Access   []TagAccess      `json:"access"`
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
	t.Memory.Lock()
	defer t.Memory.Unlock()
	t.History = append(t.History, TagHistoryItem{
		ClientID: clientID,
		Hash:     hash,
		Created:  created,
	})
}

func (t *Tag) GetHistory() []TagHistoryItem {
	t.Memory.RLock()
	defer t.Memory.RUnlock()
	return t.History
}
