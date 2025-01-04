package songdata

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
