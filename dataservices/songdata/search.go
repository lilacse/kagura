package songdata

import (
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type KeyMatchResult struct {
	Key   string
	Song  Song
	Score int
}

func (svc *Service) Search(query string, limit int) []Song {
	res := make([]Song, 0, limit)

	fullMatch, ok := svc.titleSearch(query)
	if ok {
		res = append(res, fullMatch)
		return res
	}

	return svc.keySearch(strings.ToLower(query), limit)
}

func (svc *Service) GetChartById(id int) (Chart, Song, bool) {
	chart := svc.chartIdMap[id]
	if chart.Id == 0 {
		return Chart{}, Song{}, false
	}

	songId := svc.chartSongMap[chart.Id]
	song := svc.songIdMap[songId]

	return chart, song, true
}

func (svc *Service) titleSearch(title string) (Song, bool) {
	song, ok := svc.titleMap[title]
	if !ok {
		return Song{}, false
	}

	return song, true
}

func (svc *Service) keySearch(key string, limit int) []Song {
	matchRes := make([]KeyMatchResult, 0, len(svc.keyMap))

	for _, song := range svc.data {
		for _, searchKey := range song.SearchKeys {
			currRes := KeyMatchResult{Key: searchKey, Song: svc.keyMap[searchKey]}

			atStart := true
			isNewWord := true
			wordCount := 0
			lastMatchScore := 0
			searchPos := 0
			var cont bool

			for _, k := range key {
				cont = false

				for _, s := range searchKey[searchPos:] {
					var score int

					isSeperator := s == ' ' || !(s >= 'A' && s <= 'Z' || s >= 'a' && s <= 'z' || s >= '0' && s <= '9')
					match := k == s || (unicode.IsSpace(k) && isSeperator)

					if match {
						cont = true

						if atStart {
							score = 30
						} else if lastMatchScore > 0 {
							score = lastMatchScore + 1
						} else if isNewWord {
							score = 20 - wordCount
						} else {
							score = 1
						}
					} else {
						score = 0
					}

					if isSeperator {
						if !atStart {
							isNewWord = true
							wordCount += 1
						}
					} else {
						isNewWord = false
						atStart = false
					}

					currRes.Score += score
					lastMatchScore = score
					searchPos += 1

					if match {
						break
					}
				}

				if !cont {
					currRes.Score = 0
					break
				}
			}

			matchRes = append(matchRes, currRes)
		}
	}

	slices.SortFunc(matchRes, func(a, b KeyMatchResult) int {
		diff := b.Score - a.Score
		if diff != 0 {
			return diff
		}

		// if both songs share the same title, prioritise the one earlier added to the game
		if a.Song.Title == b.Song.Title {
			verA := strings.Split(a.Song.GetSongVer(), ".")
			verB := strings.Split(b.Song.GetSongVer(), ".")

			majorA, _ := strconv.Atoi(verA[0])
			majorB, _ := strconv.Atoi(verB[0])
			majorDiff := majorA - majorB
			if majorDiff != 0 {
				return majorDiff
			}

			minorA, _ := strconv.Atoi(verA[1])
			minorB, _ := strconv.Atoi(verB[1])
			minorDiff := minorA - minorB
			if minorDiff != 0 {
				return minorDiff
			}

			patchA, _ := strconv.Atoi(verA[2])
			patchB, _ := strconv.Atoi(verB[2])
			patchDiff := patchA - patchB
			if patchDiff != 0 {
				return patchDiff
			}
		}

		return strings.Compare(a.Key, b.Key)
	})

	res := make([]Song, 0, limit)
	for i := 0; i < limit; i++ {
		m := matchRes[i]

		if m.Score == 0 {
			break
		}
		res = append(res, m.Song)
	}

	return res
}
