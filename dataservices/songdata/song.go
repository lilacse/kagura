package songdata

import "math"

type Song struct {
	Id         int      `json:"id"`
	Title      string   `json:"title"`
	AltTitle   string   `json:"altTitle"`
	Artist     string   `json:"artist"`
	Charts     []Chart  `json:"charts"`
	SearchKeys []string `json:"searchKeys"`
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
