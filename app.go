package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Application struct {
	AccessLogs []AccessLog     `json:"access_logs"`
	Gateway    *http.ServeMux  `json:"-"`
	FQDN       string          `json:"fqdn"`
	Tags       map[string]*Tag `json:"tags"`
	DB         Database        `json:"-"`
	Memory     *sync.RWMutex   `json:"-"`
}

type AccessLog struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Timestamp int    `json:"timestamp"`
	TagID     string `json:"tag_id"`
}

func NewApplication(fqdn string, db Database) *Application {
	app := &Application{
		Gateway: http.NewServeMux(),
		FQDN:    fqdn,
		Tags:    map[string]*Tag{},
		DB:      db,
		Memory:  &sync.RWMutex{},
	}
	tags, err := db.GetTags()
	if err != nil {
		log.Println("error getting tags", err)
	}
	for _, tag := range tags {
		app.Tags[tag.ID] = tag
		app.Gateway.HandleFunc(fmt.Sprintf("/%s", tag.URL), app.tagHandler(tag))
	}
	// app.Gateway.HandleFunc("/tag", app.TagHandler)
	app.Gateway.HandleFunc("/tag_exists", app.TagExistsHandler)
	app.Gateway.HandleFunc("/get_tag", app.GetTagHandler)
	return app
}

func (a *Application) AddTag(tag *Tag) {
	a.Memory.Lock()
	defer a.Memory.Unlock()
	myTag, ok := a.Tags[tag.ID]
	if !ok {
		tagFromDB, err := a.DB.GetTag(tag.ID)
		if err != nil {
			fmt.Println("adding tag")
			if err := a.DB.InsertTag(tag); err != nil {
				fmt.Println("error inserting tag", err)
				return
			}
		}
		if tagFromDB != nil {
			tag.History = append(tag.History, tagFromDB.History...)
		}
		a.Tags[tag.ID] = tag
		return
	}
	myTag.History = append(myTag.History, tag.History...)
	myTag.Hash = tag.Hash
}

func (a *Application) GetTag(id string) *Tag {
	a.Memory.RLock()
	defer a.Memory.RUnlock()
	myTag, ok := a.Tags[id]
	if !ok {
		myTag, err := a.DB.GetTag(id)
		if err != nil {
			fmt.Println("error getting tag", err)
			return nil
		}
		if myTag != nil {
			a.Tags[id] = myTag
		}
		return myTag
	}
	return myTag
}
