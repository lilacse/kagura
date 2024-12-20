package step

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/lilacse/kagura/commands"
	"github.com/lilacse/kagura/dataservices/songdata"
	"github.com/lilacse/kagura/embedbuilder"
	"github.com/lilacse/kagura/store"
)

func Handle(ctx context.Context, e *gateway.MessageCreateEvent) bool {
	params, ok := commands.ExtractParamsString("step", e.Message.Content)
	if !ok {
		return false
	}

	st := store.GetState()

	params, stepStr, ok := commands.ExtractParamForward(params, 1)
	if !ok {
		sendFormatError(st, e)
		return true
	}

	params, scoreStr, ok := commands.ExtractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, e)
		return true
	}

	params, diffStr, ok := commands.ExtractParamReverse(params, 1)
	if !ok {
		sendFormatError(st, e)
		return true
	}

	_, songStr, ok := commands.ExtractParamReverse(params, -1)
	if !ok {
		sendFormatError(st, e)
		return true
	}

	step, err := strconv.Atoi(stepStr)
	if err != nil {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid step `%s`!", stepStr)))
		return true
	}

	score, errMsg, ok := commands.ParseScore(scoreStr)
	if !ok {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(errMsg))
	}

	matchSong := songdata.Search(songStr, 1)
	if len(matchSong) == 0 {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("No matching song found for query `%s`!", songStr)))
		return true
	}

	song := matchSong[0]

	diffKey := ""
	diffName := ""
	switch strings.ToLower(diffStr) {
	case "pst", "past":
		diffKey = "pst"
		diffName = "Past (PST)"
	case "prs", "present":
		diffKey = "prs"
		diffName = "Present (PRS)"
	case "ftr", "future":
		diffKey = "ftr"
		diffName = "Future (FTR)"
	case "etr", "eternal":
		diffKey = "etr"
		diffName = "Eternal (ETR)"
	case "byd", "beyond":
		diffKey = "byd"
		diffName = "Beyond (BYD)"
	}

	if diffKey == "" {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid difficulty `%s`!", diffStr)))
		return true
	}

	cc := 0.0
	level := ""
	ver := ""
	for _, c := range song.Charts {
		if c.Diff == diffKey {
			cc = c.CC
			level = c.Level
			ver = c.Ver
			break
		}
	}

	if cc == 0.0 {
		st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Difficulty %s does not exist for the song %s!", strings.ToUpper(diffKey), song.AltTitle)))
		return true
	}

	// treat scores submitted with 3 digits to 6 digits, we append zeroes to them until it reaches 7 digits
	if score == 100 {
		score = 10000000
	} else {
		for score < 1000000 {
			score *= 10
		}
	}

	var ptt float64
	if score >= 10000000 {
		ptt = cc + 2.0
	} else if score >= 9800000 && score < 10000000 {
		ptt = cc + 1.0 + ((float64(score) - 9800000) / 200000)
	} else {
		ptt = cc + (float64(score)-9500000)/300000
		if ptt < 0.0 {
			ptt = 0.0
		}
	}

	progress := (2.45*math.Sqrt(ptt) + 2.5) * (float64(step) / 50)
	floored := math.Floor(progress*10) / 10
	formula := fmt.Sprintf("(2.45 * sqrt(%.4f) + 2.5) * (%v / 50) = **%.4f** (shown as **%.1f**)", ptt, step, progress, floored)

	embed := discord.Embed{
		Fields: []discord.EmbedField{
			{
				Name:  "Song",
				Value: fmt.Sprintf("%s - %s", song.Title, song.Artist),
			},
			{
				Name:  "Chart",
				Value: fmt.Sprintf("%s - Lv%s (%.1f) (v%s)", diffName, level, cc, ver),
			},
			{
				Name:   "Score",
				Value:  strconv.Itoa(score),
				Inline: true,
			},
			{
				Name:   "Play Rating",
				Value:  fmt.Sprintf("%.4f", ptt),
				Inline: true,
			},
			{
				Name:   "Step stat",
				Value:  strconv.Itoa(step),
				Inline: true,
			},
			{
				Name: "Progress gained",
				Value: fmt.Sprintf(`%s

-# - There might be a ±0.1 difference in actual progress gained due to differences in calculation performed by the game.
-# - For partner progression bonuses, __add__ them to the value above before calculating Play+ and fragment boosts.
-# - For Play+ boost, __multiply__ the value by stamina used. For fragment boost, further __multiply__ the value by boost multiplier.`, formula),
			},
		},
	}

	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.Info(embed))
	return true
}

func sendFormatError(st *state.State, e *gateway.MessageCreateEvent) {
	st.SendEmbedReply(e.ChannelID, e.ID, embedbuilder.UserError(fmt.Sprintf("Invalid input, expecting `%sstep [stat] [song] [diff] [score]`!", store.GetPrefix())))
}
