package main

import (
	"context"
	"fmt"
	"log"
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
			id UUID PRIMARY KEY,
			username TEXT,
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
        INSERT INTO tags (id, username, file_path, client_id, hash, created, history)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO NOTHING
    `, tag.ID, tag.Username, tag.FilePath, tag.ClientID, tag.Hash, tag.Created, tag.History)
	return err
}

func (p *PostgresDB) GetTag(id string) (*Tag, error) {
	var tag Tag
	err := p.Pool.QueryRow(context.Background(), `
		SELECT id, username, file_path, client_id, hash, created, history
		FROM tags
		WHERE id = $1
	`, id).Scan(&tag.ID, &tag.Username, &tag.FilePath, &tag.ClientID, &tag.Hash, &tag.Created, &tag.History)
	if err != nil {
		return nil, err
	}
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
		if err := rows.Scan(&tag.ID, &tag.Username, &tag.FilePath, &tag.ClientID, &tag.Hash, &tag.Created, &tag.History); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

func (p *PostgresDB) UpdateTag(tag *Tag) error {
	_, err := p.Pool.Exec(context.Background(), `
		UPDATE tags
		SET client_id = $2, hash = $3, created = $4, history = $5, username = $6, file_path = $7
		WHERE id = $1
	`, tag.ID, tag.ClientID, tag.Hash, tag.Created, tag.History, tag.Username, tag.FilePath)
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
	month := strings.ToLower(currentTime.Format("Jan"))
	year := currentTime.Format("2006")
	tableName := fmt.Sprintf("access_logs_%s_%s", month, year)

	// First, create the table if it doesn't exist
	createQuery := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id SERIAL PRIMARY KEY,
            ip TEXT,
            user_agent TEXT,
            timestamp INT,
            tag_id TEXT
        )`, tableName)
	_, err := p.Pool.Exec(context.Background(), createQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Then, insert the data
	insertQuery := fmt.Sprintf(`
        INSERT INTO %s (ip, user_agent, timestamp, tag_id)
        VALUES ($1, $2, $3, $4)`, tableName)
	_, err = p.Pool.Exec(context.Background(), insertQuery,
		log.IP,
		log.UserAgent,
		log.Timestamp,
		log.TagID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert log: %v", err)
	}

	return nil
}

func (p *PostgresDB) GetAccessLogs(table string) ([]*AccessLog, error) {
	rows, err := p.Pool.Query(context.Background(), `
		SELECT ip, user_agent, timestamp, tag_id
		FROM access_logs_%s
	`, table)
	if err != nil {
		log.Println("GetAccessLogs error getting access logs", err)
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
