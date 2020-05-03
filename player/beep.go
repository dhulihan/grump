package player

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dhulihan/grump/library"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
)

// BeepAudioPlayer is an audio player implementation that uses beep
type BeepAudioPlayer struct{}

// ap is the audio panel for the controller
type ap struct {
	sampleRate beep.SampleRate
	ctrl       *beep.Ctrl
	resampler  *beep.Resampler
	volume     *effects.Volume
	streamer   beep.StreamSeekCloser
	finished   bool
}

// newAudioPanel creates a new audio panel.
//
// count - number of times to repeat the track
func newAudioPanel(sampleRate beep.SampleRate, streamer beep.StreamSeekCloser, count int) *ap {
	ctrl := &beep.Ctrl{Streamer: beep.Loop(count, streamer)}
	resampler := beep.ResampleRatio(4, 1, ctrl)
	volume := &effects.Volume{Streamer: resampler, Base: 2}
	return &ap{
		sampleRate: sampleRate,
		ctrl:       ctrl,
		resampler:  resampler,
		volume:     volume,
		streamer:   streamer,
	}
}

func NewBeepAudioPlayer() (*BeepAudioPlayer, error) {
	bmp := BeepAudioPlayer{}
	return &bmp, nil
}

// returns an error and channel that specifies if the media is done playing
func (bmp *BeepAudioPlayer) Play(track library.Track, repeat bool) (*Controller, error) {
	c := Controller{
		path: track.Path,
		done: make(chan (bool)),
	}

	f, err := os.Open(track.Path)
	if err != nil {
		return nil, err
	}
	// do not close file io, this should get freed up when we close the streamer
	//defer f.Close()

	// assume everything is mp3 for now
	// TODO: support other formats
	// TODO: don't do this on every file, since this resets the speaker
	// WARNING: must close streamer in caller since we're not doing this here
	var s beep.StreamSeekCloser
	var format beep.Format

	switch track.FileType {
	case "MP3":
		s, format, err = mp3.Decode(f)
		if err != nil {
			return nil, err
		}
	case "FLAC":
		s, format, err = flac.Decode(f)
		if err != nil {
			return nil, err
		}
	case "WAV":
		return nil, errors.New("wav files not supported yet")
	default:
		return nil, fmt.Errorf("unsupported file type [%s]", track.FileType)
	}

	// do not close streamer, no audio will play
	//defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/30))

	// number of times to repeat the track
	count := 1
	if repeat {
		count = -1
	}

	c.ap = newAudioPanel(format.SampleRate, s, count)

	// WARNING: speaker.Play is async
	speaker.Play(beep.Seq(c.ap.volume, beep.Callback(func() {
		logrus.WithField("path", track.Path).Debug("streamer callback firing")
		c.Stop()
	})))

	return &c, nil
}

// Done returns a done channel
func (c *Controller) Done() chan (bool) {
	return c.done
}

// Progress returns the current state of playing audio.
func (c *Controller) Progress() (PlayState, error) {
	speaker.Lock()
	p := c.ap.streamer.Position()
	position := c.ap.sampleRate.D(p)
	l := c.ap.streamer.Len()
	length := c.ap.sampleRate.D(l)
	percentageComplete := float32(p) / float32(l)
	volume := c.ap.volume.Volume
	speed := c.ap.resampler.Ratio()
	finished := c.ap.finished
	speaker.Unlock()

	positionStatus := fmt.Sprintf("%v / %v", position.Round(time.Second), length.Round(time.Second))
	volumeStatus := fmt.Sprintf("%.1f", volume)
	speedStatus := fmt.Sprintf("%.3fx", speed)

	prog := PlayState{
		Progress: percentageComplete,
		Position: positionStatus,
		Volume:   volumeStatus,
		Speed:    speedStatus,
		Finished: finished,
	}
	return prog, nil
}

// PauseToggle pauses/unpauses audio. Returns true if currently paused, false if unpaused.
func (c *Controller) PauseToggle() bool {
	speaker.Lock()
	defer speaker.Unlock()

	c.ap.ctrl.Paused = !c.ap.ctrl.Paused
	return c.ap.ctrl.Paused
}

func (c *Controller) VolumeUp() {
	speaker.Lock()
	defer speaker.Unlock()

	c.ap.volume.Volume += 0.1
}

func (c *Controller) VolumeDown() {
	speaker.Lock()
	defer speaker.Unlock()

	c.ap.volume.Volume -= 0.1
}

// SpeedUp increases speed
func (c *Controller) SpeedUp() {
	speaker.Lock()
	defer speaker.Unlock()

	c.ap.resampler.SetRatio(c.ap.resampler.Ratio() * 16 / 15)
}

// SpeedDown slows down speed
func (c *Controller) SpeedDown() {
	speaker.Lock()
	defer speaker.Unlock()

	c.ap.resampler.SetRatio(c.ap.resampler.Ratio() * 15 / 16)
}

// SeekForward moves progress forward
func (c *Controller) SeekForward() error {
	speaker.Lock()
	defer speaker.Unlock()

	newPos := c.ap.streamer.Position()
	newPos += c.ap.sampleRate.N(time.Second * SeekSecs)
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= c.ap.streamer.Len() {
		newPos = c.ap.streamer.Len() - SeekSecs
	}
	if err := c.ap.streamer.Seek(newPos); err != nil {
		return fmt.Errorf("could not seek to new position [%d]: %s", newPos, err)
	}
	return nil
}

// SeekBackward moves progress backward
func (c *Controller) SeekBackward() error {
	speaker.Lock()
	defer speaker.Unlock()

	newPos := c.ap.streamer.Position()
	newPos -= c.ap.sampleRate.N(time.Second * SeekSecs)
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= c.ap.streamer.Len() {
		newPos = c.ap.streamer.Len() - 1
	}
	if err := c.ap.streamer.Seek(newPos); err != nil {
		return fmt.Errorf("could not seek to new position [%d]: %s", newPos, err)
	}
	return nil
}

// Stop must be thread safe
func (c *Controller) Stop() {
	// free up streamer
	// NOTE: this will cause the stremer to finish, and the seq callback will
	// fire
	c.ap.finished = true

	if c.ap.streamer == nil {
		return
	}

	c.ap.streamer.Close()
}
