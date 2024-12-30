package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/lilacse/kagura/logger"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type DbService struct {
	db *sql.DB
}

func NewDbService(ctx context.Context) (*DbService, error) {
	dbPath := os.Getenv("KAGURA_DBPATH")
	if dbPath == "" {
		logger.Info(ctx, "environment variable KAGURA_DBPATH is not set, using default path kagura.db")
		dbPath = "kagura.db"
	}

	uri := fmt.Sprintf("file:%s", dbPath)
	logger.Info(ctx, fmt.Sprintf("opening database on %s", uri))
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open database with uri %s", uri)
	}

	err = setupDb(db)
	if err != nil {
		return nil, err
	}

	return &DbService{db: db}, nil
}

func setupDb(db *sql.DB) error {
	ddls := []string{
		`create table if not exists scores (
			id integer primary key,
			user_id integer, 
			chart_id integer, 
			score integer, 
			timestamp integer
		)`,
		`create index if not exists scores_idx on scores (
			user_id, 
			chart_id
		)`,
		`PRAGMA journal_mode=WAL`,
		`PRAGMA synchronous=NORMAL`,
	}

	for _, ddl := range ddls {
		_, err := db.Exec(ddl)

		if err != nil {
			return fmt.Errorf("failed to setup database: %v", err)
		}
	}

	return nil
}

func (svc *DbService) NewSession(ctx context.Context) (*DbSession, error) {
	conn, err := svc.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to database: %v", err)
	}

	return &DbSession{Conn: conn}, nil
}

func (svc *DbService) Close() error {
	err := svc.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}

	return nil
}
