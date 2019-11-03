package songmem

import (
	"math"
	"time"
)

// See https://wiki.mozilla.org/User:Jesse/NewFrecency
func songHearingsToFrecentSongs(shs []songHearing) []string {
	now := time.Now()
	const lambda float64 = 0.00096270442 // (ln 2) / (30 days * 24h)

	songToFrecency := make(map[string]float64)
	for _, sh := range shs {
		hearingAge := now.Sub(sh.Date).Hours()
		songToFrecency[sh.Name] += math.Exp(-lambda * hearingAge)
	}

	return songRatingsToSongs(songToFrecency)
}
