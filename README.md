[![GoDoc](https://godoc.org/github.com/codesoap/songs?status.svg)](https://godoc.org/github.com/codesoap/songs)

# Usage
```
Usage:
    songs --register [--no-add] <name>
    songs
    songs --added-at
    songs --favourite
    songs --frecent
    songs --suggestions <name>
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
```
