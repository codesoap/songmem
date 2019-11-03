package songmem

import (
	"sort"
)

type songRating struct {
	Song   string
	Rating float64
}

// songRatingsToSongs returns a slice of songs, ordered by their rating.
func songRatingsToSongs(ratingsMap map[string]float64) []string {
	songRatings := make([]songRating, 0, len(ratingsMap))
	for song, rating := range ratingsMap {
		songRatings = append(songRatings, songRating{song, rating})
	}
	sort.Slice(songRatings, func(i, j int) bool {
		return songRatings[i].Rating > songRatings[j].Rating
	})

	songs := make([]string, 0, len(songRatings))
	for _, songRating := range songRatings {
		songs = append(songs, songRating.Song)
	}
	return songs
}
