package database

import (
	"database/sql"
)

type Session struct {
	Conn *sql.Conn
}

func (sess *Session) GetScoresRepo() *ScoresRepo {
	return GetScoresRepo(sess.Conn)
}
