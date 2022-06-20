package main

import (
	"fmt"
	"github.com/codesoap/songmem"
	"github.com/docopt/docopt-go"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var usage = `
Usage:
    songmem --register [--no-add] <name>
    songmem
    songmem --added-at
    songmem [--omit=<timespan>] --favourite
    songmem [--omit=<timespan>] --frecent
    songmem [--omit=<timespan>] --suggestions <name>
    songmem --remove-hearing [<name>]
    songmem --remove-song [<name>]
    songmem --rename <name> <newname>
Options:
    -h --help         Show this screen.
    -r --register     Register that you just heard a song. If the song does not
                      exist yet, it will be added to the database.
    -n --no-add       Do not add a song to the database, when registering that
                      you just heard it. If the song does not exist, nothing
                      will happen.
    -t --added-at     List songs by the date of their addition. Newest first.
    -f --favourite    List songs you heard the most. Most heard first.
    -c --frecent      List songs you recently heard a lot. Most frecent first.
    -s --suggestions  List songs, that you often hear before or after hearing
                      the given song. Best suggestions first.
    -o --omit=<timespan>  Exclude songs that were heard within <timespan> before
                          now. <timespan> may be something like 30m or 2h.
    --remove-hearing  Remove the latest hearing from the database. If <name> is
                      given, remove the latest hearing of the given song.
    --remove-song     Remove the last added song from the database. If <name> is
                      given, remove this song. Fails if there are still hearings
                      of the song.
    --rename          Rename the song <name> to <newname>.

If songmem is called without any arguments, it will list all songs, last heard
first.
`

type conf struct {
	Name          string
	Register      bool
	NoAdd         bool
	AddedAt       bool
	Favourite     bool
	Frecent       bool
	Suggestions   bool
	Omit          string
	RemoveHearing bool
	RemoveSong    bool
	Rename        bool
	Newname       string
}

func main() {
	opts, err := docopt.ParseDoc(usage)
	if err != nil {
		// No need to print anything here, since ParseDoc() already does.
		os.Exit(1)
	}
	var conf conf
	err = opts.Bind(&conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, `Error when using arguments:`, err.Error())
		os.Exit(2)
	}
	conf.Name = strings.TrimSpace(conf.Name)
	conf.Newname = strings.TrimSpace(conf.Newname)

	db, err := songmem.InitDB(getDBFilename())
	defer db.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, `Error when initializing database:`,
			err.Error())
		os.Exit(3)
	}
	err = db.CreateSchemaIfNotExists()
	if err != nil {
		fmt.Fprintln(os.Stderr, `Error when creating database schema:`,
			err.Error())
		os.Exit(4)
	}

	switch {
	case conf.Register && conf.NoAdd:
		err = db.AddHearing(conf.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when adding hearing:`, err.Error())
			os.Exit(5)
		}
	case conf.Register:
		sanityCheckName(conf.Name)
		err = db.AddHearingAndSongIfNeeded(conf.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when adding song or hearing:`,
				err.Error())
			os.Exit(6)
		}
	case conf.AddedAt:
		songs, err := db.ListSongsInOrderOfAddition()
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(7)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	case conf.Favourite:
		var songs []string
		if conf.Omit != "" {
			omit, err := time.ParseDuration(conf.Omit)
			if err != nil {
				errMsg := `Could not parse duration "` + conf.Omit + `":`
				fmt.Fprintln(os.Stderr, errMsg, err.Error())
				os.Exit(8)
			}
			songs, err = db.ListFavouriteSongsOmitting(omit)
		} else {
			songs, err = db.ListFavouriteSongs()
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(8)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	case conf.Frecent:
		var songs []string
		if conf.Omit != "" {
			omit, err := time.ParseDuration(conf.Omit)
			if err != nil {
				errMsg := `Could not parse duration "` + conf.Omit + `":`
				fmt.Fprintln(os.Stderr, errMsg, err.Error())
				os.Exit(9)
			}
			songs, err = db.ListFrecentSongsOmitting(omit)
		} else {
			songs, err = db.ListFrecentSongs()
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(9)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	case conf.Suggestions:
		var songs []string
		if conf.Omit != "" {
			omit, err := time.ParseDuration(conf.Omit)
			if err != nil {
				errMsg := `Could not parse duration "` + conf.Omit + `":`
				fmt.Fprintln(os.Stderr, errMsg, err.Error())
				os.Exit(10)
			}
			songs, err = db.ListSuggestionsOmitting(conf.Name, omit)
		} else {
			songs, err = db.ListSuggestions(conf.Name)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(10)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	case conf.RemoveHearing:
		song := conf.Name
		if len(conf.Name) > 0 {
			err = db.RemoveLastHearingOf(conf.Name)
		} else {
			song, err = db.RemoveLastHearing()
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when removing hearing:`, err.Error())
			os.Exit(11)
		}
		fmt.Fprintln(os.Stderr, "Removed latest hearing of:", song)
	case conf.RemoveSong:
		song := conf.Name
		if len(conf.Name) > 0 {
			err = db.RemoveSong(conf.Name)
		} else {
			song, err = db.RemoveLastAddedSong()
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when removing song:`, err.Error())
			os.Exit(12)
		}
		fmt.Fprintln(os.Stderr, "Removed song:", song)
	case conf.Rename:
		sanityCheckName(conf.Newname)
		err = db.RenameSong(conf.Name, conf.Newname)
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when renaming song:`, err.Error())
			os.Exit(13)
		}
		fmt.Fprintln(os.Stderr, "Renamed song", conf.Name, "to", conf.Newname)
	default:
		songs, err := db.ListSongsInOrderOfLastHearing()
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(14)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	}
}

func getDBFilename() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".local/share/")
	}
	return filepath.Join(dataDir, "songmem.sql")
}

func sanityCheckName(name string) {
	if len(name) == 0 {
		fmt.Fprintln(os.Stderr, `Error: The given name is empty.`)
		os.Exit(2)
	}
	if len(name) > 100 {
		fmt.Fprintln(os.Stderr, `Error: The given name is too long.`)
		os.Exit(2)
	}
	if strings.Contains(name, "\n") {
		fmt.Fprintln(os.Stderr, `Error: The given name contains a newline character.`)
		os.Exit(2)
	}
}
