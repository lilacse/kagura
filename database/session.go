package database

import (
	"database/sql"
)

type DbSession struct {
	Conn *sql.Conn
}

func (sess *DbSession) GetScoresRepo() *ScoresRepo {
	return GetScoresRepo(sess.Conn)
}
