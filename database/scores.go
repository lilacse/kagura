package database

import (
	"context"
	"database/sql"
)

type Score struct {
	Id        int64
	UserId    int64
	ChartId   int
	Score     int
	Timestamp int64
}

type ScoresRepo struct {
	conn *sql.Conn
}

func GetScoresRepo(conn *sql.Conn) *ScoresRepo {
	return &ScoresRepo{conn: conn}
}

func (repo *ScoresRepo) Insert(ctx context.Context, userId int64, chartId int, score int, timestamp int64) (sql.Result, error) {
	return repo.conn.ExecContext(
		ctx,
		`insert into scores (user_id, chart_id, score, timestamp) values (?, ?, ?, ?)`,
		userId, chartId, score, timestamp,
	)
}

func (repo *ScoresRepo) GetById(ctx context.Context, id int64) ([]Score, error) {
	rows, err := repo.conn.QueryContext(
		ctx,
		`select id, user_id, chart_id, score, timestamp from scores where id = ?`,
		id,
	)

	if err != nil {
		return nil, err
	}

	return scanToScores(rows)
}

func (repo *ScoresRepo) GetByUser(ctx context.Context, userId int64) ([]Score, error) {
	rows, err := repo.conn.QueryContext(
		ctx,
		`select id, user_id, chart_id, score, timestamp from scores where user_id = ?`,
		userId,
	)

	if err != nil {
		return nil, err
	}

	return scanToScores(rows)
}

func (repo *ScoresRepo) GetByUserAndChart(ctx context.Context, userId int64, chartId int) ([]Score, error) {
	rows, err := repo.conn.QueryContext(
		ctx,
		`select id, user_id, chart_id, score, timestamp from scores where user_id = ? and chart_id = ?`,
		userId, chartId,
	)

	if err != nil {
		return nil, err
	}

	return scanToScores(rows)
}

func (repo *ScoresRepo) GetByUserAndChartWithOffset(ctx context.Context, userId int64, chartId int, offset int, limit int) ([]Score, error) {
	rows, err := repo.conn.QueryContext(
		ctx,
		`select id, user_id, chart_id, score, timestamp from scores where user_id = ? and chart_id = ? order by timestamp desc limit ? offset ?`,
		userId, chartId, limit, offset,
	)

	if err != nil {
		return nil, err
	}

	res, err := scanToScores(rows)
	return res, err
}

func (repo *ScoresRepo) GetScoreCountByUserAndChart(ctx context.Context, userId int64, chartId int) (int, error) {
	row, err := repo.conn.QueryContext(
		ctx,
		`select count(1) from scores where user_id = ? and chart_id = ?`,
		userId, chartId,
	)

	if err != nil {
		return -1, err
	}

	var count int
	row.Next()
	row.Scan(&count)
	return count, nil
}

func (repo *ScoresRepo) GetBestScoreByUserAndChart(ctx context.Context, userId int64, chartId int) (Score, error) {
	row, err := repo.conn.QueryContext(
		ctx,
		`select id, user_id, chart_id, score, timestamp from scores where user_id = ? and chart_id = ? order by score desc limit 1`,
		userId, chartId,
	)

	if err != nil {
		return Score{}, err
	}

	res, err := scanToScores(row)
	return res[0], err
}

func (repo *ScoresRepo) Delete(ctx context.Context, id int64) (sql.Result, error) {
	return repo.conn.ExecContext(
		ctx,
		`delete from scores where id = ?`,
		id,
	)
}

func scanToScores(rows *sql.Rows) ([]Score, error) {
	res := make([]Score, 0)

	for rows.Next() {
		s := Score{}
		err := rows.Scan(&s.Id, &s.UserId, &s.ChartId, &s.Score, &s.Timestamp)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}

	return res, nil
}
