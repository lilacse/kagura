package songdata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lilacse/kagura/logger"
)

type SongData []Song

var data SongData
var titleMap map[string]Song
var keyMap map[string]Song
var chartIdMap map[int]Chart
var songIdMap map[int]Song
var chartSongMap map[int]int

const datapath = "data/songdata.json"

func Init(ctx context.Context) {
	logger.Info(ctx, "reloading song data")
	st := time.Now()

	err := loadData()
	if err != nil {
		logger.Error(ctx, err.Error())
		return
	}

	titleMap = make(map[string]Song)
	keyMap = make(map[string]Song)
	chartIdMap = make(map[int]Chart)
	songIdMap = make(map[int]Song)
	chartSongMap = make(map[int]int)

	for _, song := range data {
		titleMap[song.Title] = song
		songIdMap[song.Id] = song
		for _, key := range song.SearchKeys {
			keyMap[key] = song
		}
		for _, chart := range song.Charts {
			chartIdMap[chart.Id] = chart
			chartSongMap[chart.Id] = song.Id
		}
	}

	logger.Info(ctx, "search maps successfully rebuilt")
	logger.Info(ctx, fmt.Sprintf("song data reloaded in %s", time.Since(st)))
}

func loadData() error {
	buf, err := os.ReadFile(datapath)

	if err != nil {
		return fmt.Errorf("failed to open and read %s: %w", datapath, err)
	}

	err = json.Unmarshal(buf, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s: %s", datapath, err)
	}

	return nil
}
