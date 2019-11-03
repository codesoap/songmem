package main

import (
	"fmt"
	"github.com/codesoap/songmem"
	"github.com/docopt/docopt-go"
	"os"
	"path/filepath"
)

var usage = `
Usage:
    songmem --register [--no-add] <name>
    songmem
    songmem --added-at
    songmem --favourite
    songmem --frecent
    songmem --suggestions <name>
Options:
    -h --help         Show this screen.
    -r --register     Register that you just heard a song. If the song does not
                      exist yet, it will be added.
    -n --no-add       Do not add a song to the database, when registering that
                      you just heard it. If the song does not exist, nothing
                      will happen.
    -t --added-at     List songs by the date of their addition. Newest first.
    -f --favourite    List songs you heard the most. Most heard first.
    -c --frecent      List songs you recently heard a lot. Most frecent first.
    -s --suggestions  List songs, that you often hear before or after hearing
                      the given song. Best suggestions first.
`

type conf struct {
	Name        string
	Register    bool
	NoAdd       bool
	AddedAt     bool
	Favourite   bool
	Frecent     bool
	Suggestions bool
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
		songs, err := db.ListFavouriteSongs()
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(8)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	case conf.Frecent:
		songs, err := db.ListFrecentSongs()
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(9)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	case conf.Suggestions:
		songs, err := db.ListSuggestions(conf.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(10)
		}
		for _, s := range songs {
			fmt.Println(s)
		}
	default:
		songs, err := db.ListSongsInOrderOfLastHearing()
		if err != nil {
			fmt.Fprintln(os.Stderr, `Error when listing songs:`, err.Error())
			os.Exit(11)
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
