package player

import (
	"fmt"
	"os"
	"time"

	"github.com/dhulihan/grump/library"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	log "github.com/sirupsen/logrus"
)

var (
	speakerInitialized = false

	prevSampleRate beep.SampleRate
)

const (
	// beep quality to use for playing audio
	quality = 4
)

var (
	// maxSampleRate is used for resampling various audio formats. We also set
	// the sample rate of the speaker to this, so it essentially controls the
	// maximum quality of files played by BeepAudioPlayer.
	maxSampleRate beep.SampleRate = 44100
)

// BeepAudioPlayer is an audio player implementation that uses beep
type BeepAudioPlayer struct{}

// BeepController manages playing audio.
//
// TODO: make this an interface. this is fine for now since we're only using
// beep our audio player.
type BeepController struct {
	audioPanel *audioPanel
	path       string
	done       chan (bool)
}

// audioPanel is the audio panel for the controller
type audioPanel struct {
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
func newAudioPanel(sampleRate beep.SampleRate, streamer beep.StreamSeekCloser, count int) *audioPanel {
	ctrl := &beep.Ctrl{Streamer: beep.Loop(count, streamer)}

	log.WithFields(log.Fields{
		"src": sampleRate,
		"dst": maxSampleRate,
	}).Debug("resampling")

	resampler := beep.Resample(quality, sampleRate, maxSampleRate, ctrl)

	volume := &effects.Volume{Streamer: resampler, Base: 2}
	return &audioPanel{
		sampleRate: sampleRate,
		ctrl:       ctrl,
		resampler:  resampler,
		volume:     volume,
		streamer:   streamer,
	}
}

// NewBeepAudioPlayer --
func NewBeepAudioPlayer() (*BeepAudioPlayer, error) {
	bmp := BeepAudioPlayer{}
	return &bmp, nil
}

// Play a track and return a controller that lets you perform changes to a running track.
func (bmp *BeepAudioPlayer) Play(track library.Track, repeat bool) (AudioController, error) {
	c := BeepController{
		path: track.Path,
		done: make(chan (bool)),
	}

	f, err := os.Open(track.Path)
	if err != nil {
		return nil, err
	}
	// do not close file io, this should get freed up when we close the streamer
	//defer f.Close()

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
	case "OGG":
		s, format, err = vorbis.Decode(f)
		if err != nil {
			return nil, err
		}
	case "WAV":
		s, format, err = wav.Decode(f)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file type [%s]", track.FileType)
	}

	// number of times to repeat the track
	count := 1
	if repeat {
		count = -1
	}

	if !speakerInitialized {
		log.WithField("sampleRate", format.SampleRate).Debug("init speaker")
		speaker.Init(maxSampleRate, format.SampleRate.N(time.Second/30))
		speakerInitialized = true
	}

	c.audioPanel = newAudioPanel(format.SampleRate, s, count)

	// WARNING: speaker.Play is async
	speaker.Play(beep.Seq(c.audioPanel.volume, beep.Callback(func() {
		log.WithField("path", track.Path).Trace("streamer callback firing")
		c.Stop()
	})))

	return &c, nil
}

// Done returns a done channel
func (c *BeepController) Done() chan (bool) {
	return c.done
}

// Progress returns the current state of playing audio.
func (c *BeepController) Progress() (PlayState, error) {
	speaker.Lock()
	p := c.audioPanel.streamer.Position()
	position := c.audioPanel.sampleRate.D(p)
	l := c.audioPanel.streamer.Len()
	length := c.audioPanel.sampleRate.D(l)
	percentageComplete := float32(p) / float32(l)
	volume := c.audioPanel.volume.Volume
	speed := c.audioPanel.resampler.Ratio()
	finished := c.audioPanel.finished
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
func (c *BeepController) PauseToggle() bool {
	speaker.Lock()
	defer speaker.Unlock()

	c.audioPanel.ctrl.Paused = !c.audioPanel.ctrl.Paused
	return c.audioPanel.ctrl.Paused
}

// Paused returns current pause state
func (c *BeepController) Paused() bool {
	speaker.Lock()
	defer speaker.Unlock()

	return c.audioPanel.ctrl.Paused
}

// VolumeUp the playing track
func (c *BeepController) VolumeUp() {
	speaker.Lock()
	defer speaker.Unlock()

	c.audioPanel.volume.Volume += 0.1
}

// VolumeDown the playing track
func (c *BeepController) VolumeDown() {
	speaker.Lock()
	defer speaker.Unlock()

	c.audioPanel.volume.Volume -= 0.1
}

// SpeedUp increases speed
func (c *BeepController) SpeedUp() {
	speaker.Lock()
	defer speaker.Unlock()

	c.audioPanel.resampler.SetRatio(c.audioPanel.resampler.Ratio() * 16 / 15)
}

// SpeedDown slows down speed
func (c *BeepController) SpeedDown() {
	speaker.Lock()
	defer speaker.Unlock()

	c.audioPanel.resampler.SetRatio(c.audioPanel.resampler.Ratio() * 15 / 16)
}

// SeekForward moves progress forward
func (c *BeepController) SeekForward() error {
	speaker.Lock()
	defer speaker.Unlock()

	newPos := c.audioPanel.streamer.Position()
	newPos += c.audioPanel.sampleRate.N(time.Second * SeekSecs)
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= c.audioPanel.streamer.Len() {
		newPos = c.audioPanel.streamer.Len() - SeekSecs
	}
	if err := c.audioPanel.streamer.Seek(newPos); err != nil {
		return fmt.Errorf("could not seek to new position [%d]: %s", newPos, err)
	}
	return nil
}

// SeekBackward moves progress backward
func (c *BeepController) SeekBackward() error {
	speaker.Lock()
	defer speaker.Unlock()

	newPos := c.audioPanel.streamer.Position()
	newPos -= c.audioPanel.sampleRate.N(time.Second * SeekSecs)
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= c.audioPanel.streamer.Len() {
		newPos = c.audioPanel.streamer.Len() - 1
	}
	if err := c.audioPanel.streamer.Seek(newPos); err != nil {
		return fmt.Errorf("could not seek to new position [%d]: %s", newPos, err)
	}
	return nil
}

// Stop must be thread safe
func (c *BeepController) Stop() {
	// free up streamer
	// NOTE: this will cause the stremer to finish, and the seq callback will
	// fire
	c.audioPanel.finished = true

	if c.audioPanel.streamer != nil {
		log.Trace("closing audioPanel streamer")
		c.audioPanel.streamer.Close()
	}
}
