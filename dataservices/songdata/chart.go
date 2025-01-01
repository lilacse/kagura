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
