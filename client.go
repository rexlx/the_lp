package main

import (
	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
)

type Client struct {
	Stream quic.Stream
	Addr   string
	ID     string
}

func NewClient(stream quic.Stream, addr string) *Client {
	uid := uuid.New()
	return &Client{
		Stream: stream,
		Addr:   addr,
		ID:     uid.String(),
	}
}
