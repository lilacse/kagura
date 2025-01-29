package songdata

import "fmt"

type Chart struct {
	Id    int     `json:"id"`
	Diff  string  `json:"diff"`
	Level string  `json:"level"`
	CC    float64 `json:"cc"`
	Ver   string  `json:"ver"`
}

func (c *Chart) GetDiffDisplayName() string {
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

func (c *Chart) GetScoreRating(score int) float64 {
	if c.CC == 0.0 {
		return 0.0
	}

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
func (c *Chart) GetActualScoreRating(score int) float64 {
	ptt := c.GetScoreRating(score)
	if ptt < 0.0 {
		ptt = 0.0
	}

	return ptt
}

func (c *Chart) GetScoreRatingString(score int) string {
	if c.CC != 0.0 {
		return fmt.Sprintf("%.4f", c.GetActualScoreRating(score))
	} else {
		return "?"
	}
}

func (c *Chart) GetCCString() string {
	if c.CC != 0.0 {
		return fmt.Sprintf("%.1f", c.CC)
	} else {
		return "?"
	}
}
