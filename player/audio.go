package player

import (
	"github.com/dhulihan/grump/library"
)

const (
	// SeekSecs is the amount of seconds to skip forward or backward
	SeekSecs = 5
)

// AudioPlayer is an interface for playing audio tracks.
type AudioPlayer interface {
	Play(library.Track, bool) (AudioController, error)
}

// AudioController will control playing audio
type AudioController interface {
	Paused() bool
	PauseToggle() bool
	Progress() (PlayState, error)
	SeekForward() error
	SeekBackward() error
	SpeedUp()
	SpeedDown()
	Stop()
	VolumeUp()
	VolumeDown()
}

// PlayState represents the current state of playing audio.
type PlayState struct {
	Finished bool
	Progress float32
	Position string
	Volume   string
	Speed    string
}
