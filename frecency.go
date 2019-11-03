package songs

import (
	"math"
	"sort"
	"time"
)

type frecencyInput struct {
	Name string
	Date time.Time
}

type kv struct {
	Key string
	Val float64
}

// See https://wiki.mozilla.org/User:Jesse/NewFrecency
func frecencyInputsToSongs(fis []frecencyInput) []string {
	now := time.Now()
	const lambda float64 = 0.00096270442 // (ln 2) / (30 days * 24h)

	songIdToFrecency := make(map[string]float64)
	for _, fi := range fis {
		var hearingAge float64 = now.Sub(fi.Date).Hours()
		songIdToFrecency[fi.Name] += math.Exp(-lambda * hearingAge)
	}

	// Make songIdToFrecencyKv a slice of songs sorted by frecency:
	songIdToFrecencyKv := make([]kv, 0, len(songIdToFrecency))
	for key, val := range songIdToFrecency {
		songIdToFrecencyKv = append(songIdToFrecencyKv, kv{key, val})
	}
	sort.Slice(songIdToFrecencyKv, func(i, j int) bool {
		return songIdToFrecencyKv[i].Val > songIdToFrecencyKv[j].Val
	})

	songs := make([]string, 0, len(songIdToFrecencyKv))
	for _, kv := range songIdToFrecencyKv {
		songs = append(songs, kv.Key)
	}
	return songs
}
