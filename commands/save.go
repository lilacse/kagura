package commands

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/lilacse/kagura/database"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/logger"
	"github.com/lilacse/kagura/store"
)

type saveHandler struct {
	store    *store.Store
	db       *database.Service
	songdata *songdata.Service
}

func NewSaveHandler(store *store.Store, db *database.Service, songdata *songdata.Service) *saveHandler {
	return &saveHandler{
		store:    store,
		db:       db,
		songdata: songdata,
	}
}

func (h *saveHandler) HandleSlashCommand(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	var data *discord.CommandInteraction

	switch e.Data.(type) {
	case *discord.CommandInteraction:
		data = e.Data.(*discord.CommandInteraction)
	default:
		return false
	}

	if data.Name != "save" {
		return false
	}

	st := h.store.Bot.State()

	query := data.Options.Find("song").String()
	matched := h.songdata.Search(query, 1)
	if len(matched) == 0 {
		sendSongQueryCommandError(st, query, e)
		return true
	}

	song := matched[0]

	diffKey := data.Options.Find("diff").String()
	chart, ok := song.GetChart(diffKey)
	if !ok {
		sendDiffNotExistCommandError(st, diffKey, song.EscapedAltTitle(), e)
		return true
	}

	score, errStr, ok := parseFullScore(data.Options.Find("score").String())
	if !ok {
		sendCommandErrorReply(st, errStr, e)
		return true
	}

	newId, ts, b50Pos := saveScore(ctx, h, int64(e.Sender().ID), chart.Id, score, e)

	res := createSaveResponseEmbed(song, chart, score, newId, ts, b50Pos)
	components := createSaveButtons(int64(e.Sender().ID), chart.Id)
	sendInteractionResponse(st, res, components, e)

	return true
}

func createSaveButtons(userId int64, chartId int) []discord.TopLevelComponent {
	return []discord.TopLevelComponent{
		&discord.ActionRowComponent{
			&discord.ButtonComponent{
				Label:    "Save another score",
				CustomID: discord.ComponentID(fmt.Sprintf("%v,save,%v", userId, chartId)),
			},
		},
	}
}

func saveScore(ctx context.Context, h *saveHandler, userId int64, chartId int, score int, e *gateway.InteractionCreateEvent) (int64, time.Time, int) {
	st := h.store.Bot.State()

	sess, err := h.db.NewSession(ctx)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return 0, time.Time{}, 0
	}

	defer func() {
		err := sess.Conn.Close()
		if err != nil {
			logAndSendCommandError(ctx, st, err, e)
		}
	}()

	tx, err := sess.Conn.BeginTx(ctx, nil)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return 0, time.Time{}, 0
	}

	isCommit := false

	defer func() {
		if !isCommit {
			err := tx.Rollback()
			if err != nil {
				logAndSendCommandError(ctx, st, err, e)
			}
		}
	}()

	scoresRepo := sess.GetScoresRepo()

	userScores, err := scoresRepo.GetByUserAndChart(ctx, int64(e.Sender().ID), chartId)
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return 0, time.Time{}, 0
	}

	// we store up to the 30th recently saved score and the user's best score.

	scoreCount := len(userScores)
	if scoreCount >= 30 {
		best := userScores[0]
		for _, s := range userScores {
			if s.Score > best.Score {
				best = s
			}
		}

		slices.SortFunc(userScores, func(a, b database.ScoreRecord) int {
			return int(a.Timestamp - b.Timestamp)
		})

		if best.Timestamp != userScores[0].Timestamp {
			scoresRepo.Delete(ctx, userScores[0].Id)
		} else if scoreCount > 30 {
			scoresRepo.Delete(ctx, userScores[1].Id)
		}
	}

	ts := time.Now()

	insertRes, err := scoresRepo.Insert(ctx, int64(e.Sender().ID), chartId, score, ts.UnixMilli())
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return 0, time.Time{}, 0
	}

	newId, _ := insertRes.LastInsertId()

	err = tx.Commit()
	if err != nil {
		logAndSendCommandError(ctx, st, err, e)
		return 0, time.Time{}, 0
	}

	isCommit = true

	b50Entries, err := scoresRepo.GetBestScoresByUserWithOffset(ctx, userId, 0, 50)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error when querying b50 after saving score: %s", err.Error()))
		return newId, ts, 0
	}

	for i, entry := range b50Entries {
		if entry.Id == newId {
			return newId, ts, i + 1
		}
	}

	return newId, ts, 0
}

func createSaveResponseEmbed(song songdata.Song, chart songdata.Chart, score int, newId int64, ts time.Time, b50Pos int) discord.Embed {
	embed := discord.Embed{
		Title: "Score saved",
		Fields: []discord.EmbedField{
			{
				Name:  "Song",
				Value: fmt.Sprintf("%s - %s", song.EscapedTitle(), song.EscapedArtist()),
			},
			{
				Name:  "Chart",
				Value: fmt.Sprintf("%s - Lv%s (%s) (v%s)", chart.GetDiffDisplayName(), chart.Level, chart.GetCCString(), chart.Ver),
			},
			{
				Name:   "Score",
				Value:  strconv.Itoa(score),
				Inline: true,
			},
			{
				Name:   "Play Rating",
				Value:  chart.GetScoreRatingString(score),
				Inline: true,
			},
			{
				Name:   "Timestamp",
				Value:  fmt.Sprintf("<t:%v:R>", ts.Unix()),
				Inline: true,
			},
		},
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("Clear bonus included if play rating > 0.0.\nSend `/unsave %v` to delete this score.", newId),
		},
	}

	if b50Pos > 0 {
		embed.Description = fmt.Sprintf("**#%v in your Best-50!**", b50Pos)
	}

	return embedbuilder.Info(embed)
}

func (h *saveHandler) HandleSaveAnother(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	st := h.store.Bot.State()

	val := e.Data.(*discord.ButtonInteraction).CustomID

	params := strings.Split(string(val), ",")
	receiver := params[1]
	if receiver != "save" {
		return false
	}

	userId, _ := strconv.ParseInt(params[0], 10, 64)
	chartId, _ := strconv.Atoi(params[2])

	ccs := []discord.TopLevelComponent{
		&discord.LabelComponent{
			Label: "Score",
			Component: &discord.TextInputComponent{
				CustomID:     discord.ComponentID("save_another_score_input"),
				Style:        discord.TextInputShortStyle,
				LengthLimits: [2]int{4, 8},
				Required:     true,
				Placeholder:  "10002221",
			},
		},
	}

	sendModalResponse(st, fmt.Sprintf("%v,save_another_score,%v", userId, chartId), "Save another score", ccs, e)
	return true
}

func (h *saveHandler) HandleSaveAnotherModalSubmit(ctx context.Context, e *gateway.InteractionCreateEvent) bool {
	st := h.store.Bot.State()

	in := e.Data.(*discord.ModalInteraction)
	val := in.CustomID

	params := strings.Split(string(val), ",")
	receiver := params[1]
	if receiver != "save_another_score" {
		return false
	}

	userId, _ := strconv.ParseInt(params[0], 10, 64)
	chartId, _ := strconv.Atoi(params[2])

	chart, song, _ := h.songdata.GetChartById(chartId)

	// workaround: .Find() does not seem to work for components that are nested.
	scoreInput := in.Components[0].(*discord.LabelComponent).Component.(*discord.TextInputComponent)
	scoreValue := scoreInput.Value

	score, err, ok := parseFullScore(scoreValue)
	if !ok {
		sendInteractionResponse(st, embedbuilder.UserError(err), []discord.TopLevelComponent{}, e)
		return true
	}

	newId, ts, b50Pos := saveScore(ctx, h, userId, chartId, score, e)

	res := createSaveResponseEmbed(song, chart, score, newId, ts, b50Pos)
	components := createSaveButtons(int64(e.Sender().ID), chart.Id)
	sendInteractionResponse(st, res, components, e)

	return true
}
