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
// 1 / (timespan between a songHearing and the closest hearing of the
// given song).
//
// FIXME: Someone with a background in maths could probably come up with
//        a better algorithm.
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
		// Timespans smaller than 1 minute are likely an error and would
		// mess up the correlation:
		minTimespan = math.Max(minTimespan, 1)
		correlations[sh.Name] += 1 / minTimespan
	}

	return songRatingsToSongs(correlations), nil
}
