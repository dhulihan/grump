# grump

```
Great but
Really
Ugly
Media
Player
```

** NOT YET AVAILABLE - IN DEVELOPMENT **

A very minimal CLI audio player.

## Features

* mp3/flac support
* ID3 tag scanning

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

## Backlog (move this to github later)

* Add help/legend [MVP]
	* can add this to status box before any song is playing, version too
* add search support [HIGH]
* add ID3 tag editing support [HIGH]
* add envvars/rcfile for specifying CLI flags [HIGH]
* Shuffle mode [HIGH]
* add local wav support [HIGH]
* Maintain same speed/vol between tracks [MED]
* "flash" vol/progress/speed/etc. [MED]
* add support for event hooks [MED]
* CLI subcommands [MED] - what do we need this for?
* add support for other media sources (remote, spotify, etc.) [MED]
* "Now Playing" playlist [LOW]
* Accept file path CLI argument and start playing immediately [LOW]
* Tech Debt
	* Display full path to file in a prettier way [MED]
	* Use queuing/internal playlist to handling track stop/start/play next [LOW]
* Bugs
	* Cannot seek forward/backward on FLAC files (needs to be implemented in beep)
	* Runaway threads in OSX, SIGINT should be cleaning these up.
		* `ps aux | grep "go-build"`
