package commands

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

type scoresHandler struct {
	cmd
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewScoresHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *scoresHandler {
	return &scoresHandler{
		cmd: cmd{
			cmds: []string{"scores"},
			params: [][]param{
				{
					{
						name: "song",
					},
					{
						name: "diff",
					},
				},
			},
		},
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *scoresHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, diffStr, ok := extractParamBackwards(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	_, songStr, ok := extractParamBackwards(params, -1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	matchSong := h.songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		sendSongQueryError(st, songStr, e)
		return true
	}

	song := matchSong[0]

	diffKey, ok := parseDiffKey(diffStr)
	if !ok {
		sendInvalidDiffError(st, diffStr, e)
		return true
	}

	chart, ok := song.GetChart(diffKey)
	if !ok {
		sendDiffNotExistError(st, diffKey, song.AltTitle, e)
		return true
	}

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	scoresRepo := sess.GetScoresRepo()
	scores, err := scoresRepo.GetByUserAndChart(ctx, int64(e.Author.ID), chart.Id)
	if err != nil {
		logAndSendError(ctx, st, err, e)
		return true
	}

	if len(scores) == 0 {
		sendReply(st, embedbuilder.UserError("You don't have any scores saved for this chart!"), e)
		return true
	}

	bestScore := scores[0]

	for _, s := range scores {
		if s.Score > bestScore.Score {
			bestScore = s
		}
	}

	slices.SortFunc(scores, func(a, b database.Score) int {
		return cmp.Compare(b.Timestamp, a.Timestamp)
	})

	recents := strings.Builder{}

	for i, s := range scores {
		recents.WriteString(fmt.Sprintf("%v. %v (<t:%v:R>)\n  -# Score ID: %v\n", i+1, s.Score, s.Timestamp/1000, s.Id))
	}

	embed := discord.Embed{
		Title: fmt.Sprintf("Saved Scores for %s - %s Lv%s", song.AltTitle, chart.GetDiffDisplayName(), chart.Level),
		Fields: []discord.EmbedField{
			{
				Name:  "Best score",
				Value: fmt.Sprintf("%v (Play Rating %s)\n<t:%v:R>\n-# Score ID: %v", bestScore.Score, chart.GetScoreRatingString(bestScore.Score), bestScore.Timestamp/1000, bestScore.Id),
			},
			{
				Name:  "Recent scores",
				Value: recents.String(),
			},
		},
	}

	sendReply(st, embedbuilder.Info(embed), e)

	return true
}
