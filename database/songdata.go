package database

import (
	"context"
	"database/sql"

	"github.com/lilacse/kagura/dataservices/songdata"
)

type ChartsRepo struct {
	conn *sql.Conn
}

func GetChartsRepo(conn *sql.Conn) *ChartsRepo {
	return &ChartsRepo{conn: conn}
}

func (repo *ChartsRepo) InsertCharts(ctx context.Context, songs []songdata.Song) error {
	for _, s := range songs {
		for _, c := range s.Charts {
			_, err := repo.conn.ExecContext(
				ctx,
				`insert into charts (id, cc) values (?, ?)`,
				c.Id, c.CC,
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
