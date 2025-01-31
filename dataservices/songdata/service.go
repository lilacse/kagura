package songdata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lilacse/kagura/logger"
)

type songData []Song

type Service struct {
	data         songData
	titleMap     map[string]Song
	keyMap       map[string]Song
	chartIdMap   map[int]Chart
	songIdMap    map[int]Song
	chartSongMap map[int]int
}

const datapath = "data/songdata.json"

func NewService(ctx context.Context) (*Service, error) {
	data, err := loadData(ctx)
	if err != nil {
		return nil, err
	}

	svc := Service{
		data: data,
	}

	buildSearchMaps(ctx, &svc)

	return &svc, nil
}

func loadData(ctx context.Context) (songData, error) {
	logger.Info(ctx, "reloading song data")
	st := time.Now()

	buf, err := os.ReadFile(datapath)

	if err != nil {
		return nil, fmt.Errorf("failed to open and read %s: %w", datapath, err)
	}

	data := make(songData, 0)

	err = json.Unmarshal(buf, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %s", datapath, err)
	}

	logger.Info(ctx, fmt.Sprintf("song data reloaded in %s", time.Since(st)))

	return data, nil
}

func buildSearchMaps(ctx context.Context, svc *Service) {
	logger.Info(ctx, "rebuilding search maps")
	st := time.Now()

	svc.titleMap = make(map[string]Song)
	svc.keyMap = make(map[string]Song)
	svc.chartIdMap = make(map[int]Chart)
	svc.songIdMap = make(map[int]Song)
	svc.chartSongMap = make(map[int]int)

	for _, song := range svc.data {
		svc.titleMap[song.Title] = song
		svc.songIdMap[song.Id] = song
		for _, key := range song.SearchKeys {
			svc.keyMap[key] = song
		}
		for _, chart := range song.Charts {
			svc.chartIdMap[chart.Id] = chart
			svc.chartSongMap[chart.Id] = song.Id
		}
	}

	logger.Info(ctx, fmt.Sprintf("search maps successfully rebuilt in %s", time.Since(st)))
}
