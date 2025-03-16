package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	InsertTag(tag *Tag) error
	GetTag(id string) (*Tag, error)
	GetTags() ([]*Tag, error)
	UpdateTag(tag *Tag) error
	DeleteTag(id string) error
	AddAccessLog(log *AccessLog) error
	GetAccessLogs(table string) ([]*AccessLog, error)
}

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	p := &PostgresDB{Pool: pool}
	if err := p.createTables(); err != nil {
		return nil, err
	}
	fmt.Println("PostgresDB created, cleaning old responses")
	return p, nil
}

func (p *PostgresDB) createTables() error {
	_, err := p.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			file_path TEXT,
			client_id TEXT,
			hash TEXT,
			url TEXT,
			created INT,
			history JSONB,
			access JSONB
		);
	`)
	return err
}

func (p *PostgresDB) InsertTag(tag *Tag) error {
	_, err := p.Pool.Exec(context.Background(), `
        INSERT INTO tags (id, client_id, hash, created, history)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO NOTHING
    `, tag.ID, tag.ClientID, tag.Hash, tag.Created, tag.History)
	return err
}

func (p *PostgresDB) GetTag(id string) (*Tag, error) {
	var tag Tag
	err := p.Pool.QueryRow(context.Background(), `
		SELECT id, client_id, hash, created, history
		FROM tags
		WHERE id = $1
	`, id).Scan(&tag.ID, &tag.ClientID, &tag.Hash, &tag.Created, &tag.History)
	return &tag, err
}

func (p *PostgresDB) GetTags() ([]*Tag, error) {
	rows, err := p.Pool.Query(context.Background(), `
		SELECT id, client_id, hash, created, history
		FROM tags
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.ClientID, &tag.Hash, &tag.Created, &tag.History); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

func (p *PostgresDB) UpdateTag(tag *Tag) error {
	_, err := p.Pool.Exec(context.Background(), `
		UPDATE tags
		SET client_id = $2, hash = $3, created = $4, history = $5
		WHERE id = $1
	`, tag.ID, tag.ClientID, tag.Hash, tag.Created, tag.History)
	return err
}

func (p *PostgresDB) DeleteTag(id string) error {
	_, err := p.Pool.Exec(context.Background(), `
		DELETE FROM tags
		WHERE id = $1
	`, id)
	return err
}

func (p *PostgresDB) AddAccessLog(log *AccessLog) error {
	currentTime := time.Now()
	month := currentTime.Format("Jan")
	month = strings.ToLower(month)
	year := currentTime.Format("2006")
	_, err := p.Pool.Exec(context.Background(), `
		create table if not exists access_logs_%s_%s (
			id serial primary key,
			ip text,
			user_agent text,
			timestamp int
		);
		insert into access_logs_%s_%s (ip, user_agent, timestamp, tag_id)
		values ($1, $2, $3, $4)
	`, month, year, month, year, log.IP, log.UserAgent, log.Timestamp, log.TagID)
	return err
}

func (p *PostgresDB) GetAccessLogs(table string) ([]*AccessLog, error) {
	rows, err := p.Pool.Query(context.Background(), `
		SELECT ip, user_agent, timestamp, tag_id
		FROM access_logs_%s
	`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*AccessLog
	for rows.Next() {
		var log AccessLog
		if err := rows.Scan(&log.IP, &log.UserAgent, &log.Timestamp, &log.TagID); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	return logs, nil
}
