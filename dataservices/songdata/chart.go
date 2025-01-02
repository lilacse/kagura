package songdata

type Chart struct {
	Id    int     `json:"id"`
	Diff  string  `json:"diff"`
	Level string  `json:"level"`
	CC    float64 `json:"cc"`
	Ver   string  `json:"ver"`
}

func (c *Chart) DiffDisplayName() string {
	switch c.Diff {
	case "pst":
		return "Past (PST)"
	case "prs":
		return "Present (PRS)"
	case "ftr":
		return "Future (FTR)"
	case "etr":
		return "Eternal (ETR)"
	case "byd":
		return "Beyond (BYD)"
	default:
		return ""
	}
}

func (c *Chart) ScoreRating(score int) float64 {
	var ptt float64
	if score >= 10000000 {
		ptt = c.CC + 2.0
	} else if score >= 9800000 && score < 10000000 {
		ptt = c.CC + 1.0 + ((float64(score) - 9800000) / 200000)
	} else {
		ptt = c.CC + (float64(score)-9500000)/300000
	}
	return ptt
}

// Similar to ScoreRating, but automatically converts negative rating to 0.0.
func (c *Chart) ActualScoreRating(score int) float64 {
	ptt := c.ScoreRating(score)
	if ptt < 0.0 {
		ptt = 0.0
	}

	return ptt
}
