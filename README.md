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
# build for linux (linux host)
./scripts/build-linux.sh

# build for linux (non-linux host)
docker-compose run build

# build for darwin (darwin host)
./scripts/build-darwin.sh
```

### Releasing

```sh

```
