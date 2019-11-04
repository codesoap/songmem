package songmem

import (
	"errors"
	"math"
	"time"
)

// songHearingsToSuggestions transforms []songHearing to a slice of
// songs. The songs will be ordered by their correlation to the given
// song.
//
// The algorithm for determining the correlation calculates the sum of
// e ^ (-λ_1h * abs(time_of_hearing - closest_hearing_of_given_song))
// for every song.
func songHearingsToSuggestions(shs []songHearing, song string) ([]string, error) {
	var gshts []time.Time // given song hearing times
	for _, sh := range shs {
		if sh.Name == song {
			gshts = append(gshts, sh.Date)
		}
	}
	if len(gshts) == 0 {
		return nil, errors.New("the given song was never heard")
	}

	const lambda float64 = 0.01155245301 // ln(2) / 60min
	correlations := make(map[string]float64)
	for _, sh := range shs {
		if sh.Name == song {
			continue
		}
		var minTimespan = math.MaxFloat64
		for _, gsht := range gshts {
			timespan := math.Abs(gsht.Sub(sh.Date).Minutes())
			minTimespan = math.Min(timespan, minTimespan)
		}
		correlations[sh.Name] += math.Exp(-lambda * minTimespan)
	}

	return songRatingsToSongs(correlations), nil
}
