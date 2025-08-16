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

func (sess *Session) GetChartsRepo() *ChartsRepo {
	return GetChartsRepo(sess.Conn)
}
