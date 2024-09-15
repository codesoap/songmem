package songmem

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkSongHearingsToSuggestions(b *testing.B) {
	for hearingsCnt := 1_000; hearingsCnt <= 1_000_000; hearingsCnt *= 10 {
		songCnt := hearingsCnt / 4
		b.Run(fmt.Sprintf("%d hearings of %d songs", hearingsCnt, songCnt),
			func(b *testing.B) {
				songHearings := generateSongHearings(hearingsCnt, songCnt)
				ref := songHearings[0].Name
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := songHearingsToSuggestions(songHearings, ref)
					if err != nil {
						b.Fatalf("Could not transform song hearings to suggestions: %v", err)
					}
				}
			})
	}
}

func generateSongHearings(hearingsCnt, songCnt int) []songHearing {
	song := 0
	t := time.Now()
	shs := make([]songHearing, hearingsCnt)
	for i := 0; i < hearingsCnt; i++ {
		song = (song + 1) % songCnt
		shs[i] = songHearing{
			Name: fmt.Sprint("song", i),
			Date: t.Add(time.Duration(i) * time.Second),
		}
	}
	return shs
}
