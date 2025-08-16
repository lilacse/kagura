package database

import (
	"context"
	"database/sql"
)

type ScoreRecord struct {
	Id        int64
	UserId    int64
	ChartId   int
	Score     int
	Timestamp int64
}

type ScoreRecordRating struct {
	ScoreRecord
	Rating float64
}

type ScoresRepo struct {
	conn *sql.Conn
}

const SCORE_RATING_QUERY string = `select
		best.id,
		best.user_id,
		best.chart_id,
		best.score,
		best.timestamp,
		case 
			when best.score < 9800000 then max(charts.cc + (cast(best.score as float)-9500000)/ 300000, 0)
			when best.score < 10000000 then charts.cc + 1 + (cast(best.score as float)-9800000)/ 200000
			when best.score >= 10000000 then charts.cc + 2
		end rating
	from
		(
		select
			row_number() over (partition by chart_id
		order by
			score desc) score_order,
			id,
			user_id,
			chart_id,
			score,
			timestamp
		from
			scores
		where
			user_id = ?
	) best
	inner join charts on
		best.chart_id = charts.id
	where
		best.score_order = 1
	order by 
		rating desc
	limit ?
	offset ?`

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

func (repo *ScoresRepo) GetById(ctx context.Context, id int64) ([]ScoreRecord, error) {
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

func (repo *ScoresRepo) GetByUserAndChart(ctx context.Context, userId int64, chartId int) ([]ScoreRecord, error) {
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

func (repo *ScoresRepo) GetByUserAndChartWithOffset(ctx context.Context, userId int64, chartId int, offset int, limit int) ([]ScoreRecord, error) {
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

func (repo *ScoresRepo) GetBestScoreByUserAndChart(ctx context.Context, userId int64, chartId int) (ScoreRecord, error) {
	row, err := repo.conn.QueryContext(
		ctx,
		`select id, user_id, chart_id, score, timestamp from scores where user_id = ? and chart_id = ? order by score desc limit 1`,
		userId, chartId,
	)

	if err != nil {
		return ScoreRecord{}, err
	}

	res, err := scanToScores(row)
	return res[0], err
}

func (repo *ScoresRepo) GetBestScoresByUserWithOffset(ctx context.Context, userId int64, offset int, limit int) ([]ScoreRecordRating, error) {
	rows, err := repo.conn.QueryContext(
		ctx,
		SCORE_RATING_QUERY,
		userId, limit, offset,
	)

	if err != nil {
		return nil, err
	}

	return scanToScoreRatings(rows)
}

func (repo *ScoresRepo) GetBestScoreRatingsAverage(ctx context.Context, userId int64, limit int) (float64, float64, error) {
	res, err := repo.conn.QueryContext(
		ctx,
		`select avg(rating), avg(score) from (`+SCORE_RATING_QUERY+`)`,
		userId, limit, 0,
	)

	if err != nil {
		return -1, -1, err
	}

	var avgRt float64
	var avgScore float64
	res.Next()
	res.Scan(&avgRt, &avgScore)

	return avgRt, avgScore, nil
}

func (repo *ScoresRepo) GetUserPlayedChartCount(ctx context.Context, userId int64) (int, error) {
	res, err := repo.conn.QueryContext(
		ctx,
		`select count(distinct chart_id) from scores where user_id = ?`,
		userId,
	)

	if err != nil {
		return -1, err
	}

	var count int
	res.Next()
	res.Scan(&count)

	return count, nil
}

func (repo *ScoresRepo) Delete(ctx context.Context, id int64) (sql.Result, error) {
	return repo.conn.ExecContext(
		ctx,
		`delete from scores where id = ?`,
		id,
	)
}

func scanToScores(rows *sql.Rows) ([]ScoreRecord, error) {
	res := make([]ScoreRecord, 0)

	for rows.Next() {
		s := ScoreRecord{}
		err := rows.Scan(&s.Id, &s.UserId, &s.ChartId, &s.Score, &s.Timestamp)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}

	return res, nil
}

func scanToScoreRatings(rows *sql.Rows) ([]ScoreRecordRating, error) {
	res := make([]ScoreRecordRating, 0)

	for rows.Next() {
		s := ScoreRecordRating{}
		err := rows.Scan(&s.Id, &s.UserId, &s.ChartId, &s.Score, &s.Timestamp, &s.Rating)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}

	return res, nil
}
