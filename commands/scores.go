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
	store *store.Store
	db    *database.DbService
}

func NewScoresHandler(store *store.Store, db *database.DbService) *scoresHandler {
	return &scoresHandler{
		cmd: cmd{
			cmds: []string{"scores"},
			params: []param{
				{
					name: "song",
				},
				{
					name: "diff",
				},
			},
		},
		store: store,
		db:    db,
	}
}

func (h *scoresHandler) Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := extractParamsString(h.cmds[0], e.Message.Content, h.store.Bot.Prefix())
	if !ok {
		return false
	}

	st := h.store.Bot.State()

	params, diffStr, ok := extractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	_, songStr, ok := extractParamReverse(params, -1)
	if !ok {
		sendFormatError(st, h.store.Bot.Prefix(), h.cmd, e)
		return true
	}

	matchSong := songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		sendSongQueryError(st, songStr, e)
		return true
	}

	song := matchSong[0]

	diffKey, ok := getDiffKey(diffStr)
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
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError("You don't have any scores saved for this chart!"))
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
		recents.WriteString(fmt.Sprintf("%v. %v (<t:%v:R>)\n", i+1, s.Score, s.Timestamp/1000))
	}

	embed := discord.Embed{
		Title: fmt.Sprintf("Saved Scores for %s - %s Lv%s", song.AltTitle, chart.GetDiffDisplayName(), chart.Level),
		Fields: []discord.EmbedField{
			{
				Name:  "Best score",
				Value: fmt.Sprintf("%v (Play Rating %.4f)\n<t:%v:R>", bestScore.Score, chart.GetActualScoreRating(bestScore.Score), bestScore.Timestamp/1000),
			},
			{
				Name:  "Recent scores",
				Value: recents.String(),
			},
		},
	}

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Info(embed))
	return true
}