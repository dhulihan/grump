package player

import (
	"github.com/dhulihan/grump/library"
)

// MockAudioPlayer is an audio player implementation that uses beep
type MockAudioPlayer struct{}

// NewMockAudioPlayer --
func NewMockAudioPlayer() *MockAudioPlayer {
	bmp := MockAudioPlayer{}
	return &bmp
}

// Play a track and return a controller that lets you perform changes to a running track.
func (bmp *MockAudioPlayer) Play(track library.Track, repeat bool) (AudioController, error) {
	return &MockAudioController{}, nil
}

type MockAudioController struct{}

func (p *MockAudioController) Paused() bool                 { return false }
func (p *MockAudioController) PauseToggle() bool            { return true }
func (p *MockAudioController) Progress() (PlayState, error) { return PlayState{}, nil }
func (p *MockAudioController) SeekForward() error           { return nil }
func (p *MockAudioController) SeekBackward() error          { return nil }
func (p *MockAudioController) SpeedUp()                     {}
func (p *MockAudioController) SpeedDown()                   {}
func (p *MockAudioController) Stop()                        {}
func (p *MockAudioController) VolumeUp()                    {}
func (p *MockAudioController) VolumeDown()                  {}
