package songdata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lilacse/kagura/logger"
)

type Song struct {
	Id         int      `json:"id"`
	Title      string   `json:"title"`
	AltTitle   string   `json:"altTitle"`
	Artist     string   `json:"artist"`
	Charts     []Chart  `json:"charts"`
	SearchKeys []string `json:"searchKeys"`
}

type Chart struct {
	Id    int     `json:"id"`
	Diff  string  `json:"diff"`
	Level string  `json:"level"`
	CC    float64 `json:"cc"`
	Ver   string  `json:"ver"`
}

type SongData []Song

var data SongData
var titleMap map[string]Song
var keyMap map[string]Song

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

	for _, song := range data {
		titleMap[song.Title] = song
		for _, key := range song.SearchKeys {
			keyMap[key] = song
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
