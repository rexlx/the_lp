package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

var (
	dbLocation = flag.String("db", "user=rxlx password=FOO host=192.168.86.120 dbname=tags", "Database location")
)

const (
	Protocol = "heckfire"
)

type Application struct {
	TLSConfig            *tls.Config        `json:"-"`
	QUICServer           *http3.Server      `json:"-"`
	UDPListener          *quic.Listener     `json:"-"`
	Memory               *sync.RWMutex      `json:"-"`
	Gateway              *http.ServeMux     `json:"-"`
	FQDN                 string             `json:"fqdn"`
	AccessFlushFrequency int                `json:"access_flush_frequency"`
	DB                   Database           `json:"-"`
	Tags                 map[string]*Tag    `json:"tags"`
	Clients              map[string]*Client `json:"-"`
	AccessLogs           []AccessLog        `json:"access_logs"`
	Broadcast            chan []byte        `json:"-"`
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
		fmt.Println(tag.URL, tag.ID)
		app.Gateway.HandleFunc(fmt.Sprintf("/%s", tag.ID), app.tagHandler(tag))
	}
	// app.Gateway.HandleFunc("/tag", app.TagHandler)
	app.Gateway.HandleFunc("/tag-exists", app.TagExistsHandler)
	app.Gateway.HandleFunc("/get-tag", app.GetTagHandler)
	app.Gateway.HandleFunc("/tag", app.AddTagHandler)
	app.Gateway.HandleFunc("/access", app.AccessHandler)
	return app
}

func CreateTLSConfig(certPath, keyPath, caCert string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	caCertData, err := os.ReadFile(caCert)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertData)
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		NextProtos:   []string{Protocol},
	}, nil
}

func (a *Application) AddTag(tag *Tag) {
	a.Memory.Lock()
	defer a.Memory.Unlock()
	myTag, ok := a.Tags[tag.ID]
	// not in memory
	if !ok {
		tagFromDB, err := a.DB.GetTag(tag.ID)
		// not in db, this tag is new and its safe to add a new route
		if err != nil {
			fmt.Println("adding tag")
			tag.URL = fmt.Sprintf("%s/%s", a.FQDN, tag.ID)
			a.Gateway.HandleFunc(fmt.Sprintf("/%s", tag.ID), a.tagHandler(tag))
		}
		// err is nil, tag is in db
		if tagFromDB != nil {
			tag.History = append(tag.History, tagFromDB.History...)
		}
		tag.AddHistory(tag.ClientID, tag.Hash, tag.Created)
		// store in memory
		a.Tags[tag.ID] = tag
		if err := a.DB.InsertTag(tag); err != nil {
			fmt.Println("error inserting tag", err)
		}
		return
	}
	myTag.AddHistory(tag.ClientID, tag.Hash, tag.Created)
	myTag.Hash = tag.Hash
	myTag.Created = tag.Created
	if myTag.URL == "" {
		myTag.URL = fmt.Sprintf("%s/%s", a.FQDN, tag.ID)
	}
	// a.Gateway.HandleFunc(fmt.Sprintf("/%s", myTag.URL), a.tagHandler(myTag)) // this should exist already
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

func (a *Application) AddAccess(access *AccessLog) {
	a.Memory.Lock()
	defer a.Memory.Unlock()
	a.AccessLogs = append(a.AccessLogs, *access)
	err := a.DB.AddAccessLog(access)
	if err != nil {
		fmt.Println("error adding access log", err)
	}
}

func (a *Application) handleSession(session quic.Connection) {
	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		fmt.Println("error accepting stream", err)
		return
	}
	client := NewClient(stream, session.RemoteAddr().String())
	a.Memory.Lock()
	a.Clients[client.ID] = client
	a.Memory.Unlock()
	defer func() {
		a.Memory.Lock()
		delete(a.Clients, client.ID)
		a.Memory.Unlock()
		stream.Close()
	}()
	<-stream.Context().Done()
}

func (s *Application) broadcastToClients(msg []byte) {
	s.Memory.RLock()
	defer s.Memory.RUnlock()
	for _, client := range s.Clients {
		client.Stream.Write(msg)
	}
}
