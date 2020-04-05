package player

import (
	"github.com/dhulihan/ghoulish/library"
)

const (
	// SeekSecs is the amount of seconds to skip forward or backward
	SeekSecs = 5
)

// AudioPlayer is an interface for playing audio tracks.
type AudioPlayer interface {
	Play(library.Track, bool) (*Controller, error)
}

// PlayState represents the current state of playing audio.
type PlayState struct {
	Finished bool
	Progress float32
	Position string
	Volume   string
	Speed    string
}

// Controller manages playing audio.
//
// TODO: make this an interface. this is fine for now since we're only using
// beep our audio player.
type Controller struct {
	ap   *ap
	path string
	done chan (bool)
}
