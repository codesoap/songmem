package songs

import (
	"math"
	"sort"
	"time"
)

type kv struct {
	Key string
	Val float64
}

// See https://wiki.mozilla.org/User:Jesse/NewFrecency
func songHearingsToFrecentSongs(shs []songHearing) []string {
	now := time.Now()
	const lambda float64 = 0.00096270442 // (ln 2) / (30 days * 24h)

	songIDToFrecency := make(map[string]float64)
	for _, sh := range shs {
		hearingAge := now.Sub(sh.Date).Hours()
		songIDToFrecency[sh.Name] += math.Exp(-lambda * hearingAge)
	}

	// Make songIDToFrecencyKv a slice of songs sorted by frecency:
	songIDToFrecencyKv := make([]kv, 0, len(songIDToFrecency))
	for key, val := range songIDToFrecency {
		songIDToFrecencyKv = append(songIDToFrecencyKv, kv{key, val})
	}
	sort.Slice(songIDToFrecencyKv, func(i, j int) bool {
		return songIDToFrecencyKv[i].Val > songIDToFrecencyKv[j].Val
	})

	songs := make([]string, 0, len(songIDToFrecencyKv))
	for _, kv := range songIDToFrecencyKv {
		songs = append(songs, kv.Key)
	}
	return songs
}
