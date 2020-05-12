# grump

![](screenshot.png)

```
Great but
Really
Ugly
Media
Player
```

A very minimal CLI audio player.

## Features

* cross-platform
* ID3 tag scanning
* Supports
	* FLAC
	* MP3
	* OGG/Vorbis
	* WAV

## Install

### Linux

```sh
apt install libasound2-dev build-essential
go get github.com/dhulihan/grump
```

### Mac OSX

```sh
go get github.com/dhulihan/grump
```

## Usage

```
grump path/to/some/audio/files
```

## Development

### Building

```sh
docker-compose run build
```

## Backlog (move this to github later)

* add search support [HIGH]
* sort by title/name/track [HIGH]
* add ID3 tag editing support [HIGH]
* add cli flags for log level, etc. [HIGH]
* add envvars/rcfile for specifying CLI flags [HIGH]
* Shuffle mode [HIGH]
* Maintain same speed/vol between tracks [MED]
* "flash" vol/progress/speed/etc. [MED]
* add support for event hooks [MED]
* CLI subcommands [MED] - what do we need this for?
* add support for other media sources (remote, spotify, etc.) [MED]
* customize columns [MED]
* "Now Playing" playlist [LOW]
* Accept file path CLI argument and start playing immediately [LOW]
* Tech Debt
	* Display full path to file in a prettier way [MED]
	* Use queuing/internal playlist to handling track stop/start/play next [LOW]
* Bugs
	* Cannot seek forward/backward on FLAC files (needs to be implemented in beep)
	* Runaway threads in OSX, SIGINT should be cleaning these up.
		* `ps aux | grep "go-build"`
