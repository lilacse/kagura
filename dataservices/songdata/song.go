package songdata

import (
	"math"
	"strings"
)

type Song struct {
	Id         int               `json:"id"`
	Title      string            `json:"title"`
	AltTitle   string            `json:"altTitle"`
	Artist     string            `json:"artist"`
	Charts     []Chart           `json:"charts"`
	SearchKeys []string          `json:"searchKeys"`
	Urls       map[string]string `json:"urls"`
}

func (s *Song) GetChart(diffKey string) (Chart, bool) {
	for _, c := range s.Charts {
		if c.Diff == diffKey {
			return c, true
		}
	}

	return Chart{}, false
}

func (s *Song) GetSongVer() string {
	oldest := Chart{
		Id: math.MaxInt,
	}

	for _, c := range s.Charts {
		if c.Id < oldest.Id {
			oldest = c
		}
	}

	return oldest.Ver
}

func (s *Song) EscapedTitle() string {
	return unformatString(s.Title)
}

func (s *Song) EscapedAltTitle() string {
	return unformatString(s.AltTitle)
}

func (s *Song) EscapedArtist() string {
	return unformatString(s.Artist)
}

func unformatString(t string) string {
	t = strings.ReplaceAll(t, "_", "\\_")
	t = strings.ReplaceAll(t, "*", "\\*")
	t = strings.ReplaceAll(t, "~", "\\~")
	t = strings.ReplaceAll(t, "#", "\\#")
	return t
}
