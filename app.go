package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	dbLocation = flag.String("db", "DSN", "Database location")
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
		if tag.URL == "" {
			tag.URL = fmt.Sprintf("%s/%s", app.FQDN, tag.ID)
		}
		app.Gateway.HandleFunc(fmt.Sprintf("/%s", tag.URL), app.tagHandler(tag))
	}
	// app.Gateway.HandleFunc("/tag", app.TagHandler)
	app.Gateway.HandleFunc("/tag-exists", app.TagExistsHandler)
	app.Gateway.HandleFunc("/get-tag", app.GetTagHandler)
	app.Gateway.HandleFunc("/tag", app.AddTagHandler)
	app.Gateway.HandleFunc("/access", app.AccessHandler)
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
		}
		if tagFromDB != nil {
			tag.History = append(tag.History, tagFromDB.History...)
		}
		tag.URL = fmt.Sprintf("%s/%s", a.FQDN, tag.ID)
		a.Gateway.HandleFunc(fmt.Sprintf("/%s", tag.URL), a.tagHandler(tag))
		a.Tags[tag.ID] = tag
		if err := a.DB.InsertTag(tag); err != nil {
			fmt.Println("error inserting tag", err)
		}
		return
	}
	myTag.History = append(myTag.History, TagHistoryItem{ClientID: tag.ID, Hash: tag.Hash, Created: tag.Created})
	myTag.Hash = tag.Hash
	myTag.Created = tag.Created
	if myTag.URL == "" {
		myTag.URL = fmt.Sprintf("%s/%s", a.FQDN, tag.ID)
	}
	a.Gateway.HandleFunc(fmt.Sprintf("/%s", myTag.URL), a.tagHandler(myTag))
	if err := a.DB.InsertTag(myTag); err != nil {
		fmt.Println("error inserting tag", err)
	}
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
