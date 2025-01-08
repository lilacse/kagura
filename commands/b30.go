package commands

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type b30Handler struct {
	cmd
	store *store.Store
	db    *database.DbService
}

type b30Entry struct {
	chart     songdata.Chart
	song      songdata.Song
	score     int
	rating    float64
	timestamp int64
}

func NewB30Handler(store *store.Store, db *database.DbService) *b30Handler {
	return &b30Handler{
		cmd: cmd{
			cmds:   []string{"b30"},
			params: []param{},
		},
		store: store,
		db:    db,
	}
}

func (h *b30Handler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	_, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	scoresRepo := sess.GetScoresRepo()
	scores, err := scoresRepo.GetByUser(ctx, int64(e.Author.ID))
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	chartBestMap := make(map[int]database.Score)

	for _, s := range scores {
		currBest := chartBestMap[s.ChartId]
		if s.Score > currBest.Score {
			chartBestMap[s.ChartId] = s
		}
	}

	b30Entries := make([]b30Entry, 0, len(chartBestMap))

	for s := range maps.Values(chartBestMap) {
		chart, song, ok := songdata.GetChartById(s.ChartId)
		if !ok {
			logAndSendError(ctx, st, fmt.Errorf("chart id %v is not found in songdata", s.ChartId), e)
			return true
		}

		b30Entries = append(b30Entries, b30Entry{
			chart:     chart,
			song:      song,
			score:     s.Score,
			rating:    chart.GetActualScoreRating(s.Score),
			timestamp: s.Timestamp,
		})
	}

	slices.SortFunc(b30Entries, func(a, b b30Entry) int {
		return cmp.Compare(b.rating, a.rating)
	})

	res := strings.Builder{}
	scoreCount := len(b30Entries)
	if scoreCount > 30 {
		scoreCount = 30
	}

	for i, e := range b30Entries[0:scoreCount] {
		res.WriteString(fmt.Sprintf(
			"`%02d` - %s - %s Lv%s (%.1f) - %v - **%.4f**\n",
			i+1,
			e.song.AltTitle,
			e.chart.GetDiffDisplayName(),
			e.chart.Level,
			e.chart.CC,
			e.score,
			e.rating,
		))
	}

	embed := discord.Embed{
		Title:       "Highest 30 Play Ratings from Saved Scores",
		Description: res.String(),
	}

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Info(embed))
	return true
}
