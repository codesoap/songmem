[![GoDoc](https://godoc.org/github.com/codesoap/songmem?status.svg)](https://godoc.org/github.com/codesoap/songmem)

# Installation
Installation has been tested on OpenBSD and Ubuntu, but will probably
run on any POSIX-compliant operating system, that is supported by
golang.

To install songmem run `go get -v -u 'github.com/codesoap/songmem/...'`
(this will take some time, because the dependency go-sqlite3 is
big). The songmem binary will be placed at `$HOME/go/bin/songmem`. Add
`$HOME/go/bin/` to your `$PATH` for easy usage.

# Usage
```
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
```

# Music player integration
The following scripts assume that you store your songs like
`<artist> - <title>`.

## mpd (requires mpc)
This script registers when a song is played through mpd. Start it after
launching `mpd(1)`.

```bash
#!/usr/bin/env sh

while true
do
	song="$(mpc -f '%artist% - %title%' current --wait)"
	songmem --register "$song"
done
```

You can search for recently heard songs and play them in mpd (requires
dmenu; alternatively you could use fzf):

```bash
#!/usr/bin/env sh

# Abort when dmenu is quit using <esc>:
set -e

song="$(songmem | dmenu -i -l 15 -p "Play song:")"
artist="$(printf '%s' "$song" | awk -F ' - ' '{print $1}')"
title="$(printf '%s' "$song" | awk '{i=index($0, " - "); print substr($0, i+3)}')"
songfile="$(mpc search artist "$artist" title "$title")"
if [ -n "$songfile" ]
then
	mpc clear
	mpc add "$songfile"
	mpc play
fi
```

Adapt these scripts to add songs to the queue, browse through song
suggestions for the currently playing song, ...

Setting up keyboard shortcuts for your scripts could also prove useful.

## [ytools](https://github.com/codesoap/ytools)
```bash
#!/usr/bin/env sh

# Abort when dmenu is quit using <esc> or the song cannot be played:
set -e

song="$(songmem | dmenu -i -l 15 -p "Play song:")"
ytools-search "$song"
mpv --ytdl-format "bestaudio/best" --no-video "$(ytools-pick 1)"
songmem --register "$song"
```

You could add this as a fallback into the previous script for mpd.
