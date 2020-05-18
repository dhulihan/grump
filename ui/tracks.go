package ui

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

const (
	columnStatus = iota
	columnArtist
	columnAlbum
	columnTrack

	trackIconEmptyText   = "  "
	trackIconPlayingText = "🔈"
	trackIconPausedText  = "🔇"

	// check audio progess at this interval
	checkAudioMillis = 500
)

// TrackPage is a page that displays playable audio tracks
type TrackPage struct {
	tracks                     []library.Track
	player                     player.AudioPlayer
	currentlyPlayingController *player.Controller
	currentlyPlayingTrack      *library.Track
	currentlyPlayingRow        int

	// layout
	left        *tview.List
	center      *tview.Flex
	logBox      *tview.TextView
	trackBox    *tview.Table
	progressBox *tview.Table

	theme *tview.Theme
}

// NewTrackPage generates the track page
func NewTrackPage(ctx context.Context, ml library.AudioShelf, pl player.AudioPlayer) *TrackPage {
	theme := defaultTheme()

	// Create the basic objects.
	trackBox := tview.NewTable().SetBorders(true).SetBordersColor(theme.BorderColor)

	//logBox := tview.NewBox().SetBorder(true).SetBorderColor(theme.BorderColor)
	logBox := tview.NewTextView().
		SetTextColor(theme.BorderColor)

	progressBox := tview.NewTable()
	progressBox.SetBorder(true).SetBorderColor(theme.BorderColor)

	p := &TrackPage{
		tracks:      ml.Tracks(),
		player:      pl,
		logBox:      logBox,
		trackBox:    trackBox,
		progressBox: progressBox,
		theme:       theme,
	}

	// hook our logger up to logBox so we can see log messages onscreen
	log.SetOutput(logBox)
	log.SetFormatter(p)

	return p
}

// Page populates the layout for the track page
func (t *TrackPage) Page(ctx context.Context) tview.Primitive {
	t.trackColumns(t.trackBox)

	for i, track := range t.tracks {
		// incr by one to pass table headers
		t.trackCell(t.trackBox, i+1, track)
	}

	t.trackBox.
		// fired on Escape, Tab, or Backtab key
		SetDoneFunc(func(key tcell.Key) {
			log.Debugf("done func firing, key [%v]", key)
		}).
		SetSelectable(true, false).SetSelectedFunc(t.cellChosen).SetSelectedStyle(t.theme.SecondaryTextColor, t.theme.PrimitiveBackgroundColor, tcell.AttrNone)
	//t.trackBox.SetSelectedStyle(t.theme.TertiaryTextColor, t.theme.PrimitiveBackgroundColor, tcell.AttrNone)

	t.trackBox.SetInputCapture(t.inputCapture)

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.trackBox, 0, 3, true).
		AddItem(t.progressBox, 6, 1, false).
		AddItem(t.logBox, 1, 1, false)

	// Create the layout.
	flex := tview.NewFlex().
		AddItem(main, 0, 3, true)

	t.welcome()

	// one outstanding goroutine that tracks audio progress
	go t.audioPlaying(ctx)

	return flex
}

// Format is a custom log formatter that allows write logrus entries to the ui
func (t *TrackPage) Format(entry *log.Entry) ([]byte, error) {
	// clear out the log box before writing text to it
	t.logBox.Clear()

	lf := &log.TextFormatter{
		DisableTimestamp: true,
	}
	return lf.Format(entry)
}

// main key input handler for this page
func (t *TrackPage) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	// placeholder nil check for convenience
	log.Debugf("input capture firing, name [%s] key [%d] rune [%s]", event.Name(), event.Key(), string(event.Rune()))

	// something is currently playing, handle that
	if t.currentlyPlayingController != nil {
		return t.currentlyPlayingInputCapture(event)
	}

	switch event.Key() {
	case tcell.KeyRune:
		// attempt to use rune as string
		s := string(event.Rune())
		switch s {
		case "?":
			pages.SwitchToPage("help")
		case "q":
			log.Info("exiting")
			app.Stop()
		}
	}

	return event
}

// handle key input while a track is playing
func (t *TrackPage) currentlyPlayingInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyESC:
		t.stopCurrentlyPlaying()
		t.welcome()
	case tcell.KeyLeft:
		err := t.currentlyPlayingController.SeekBackward()
		if err != nil {
			log.WithError(err).Error("problem seeking forward")
			return event
		}
	case tcell.KeyRight:
		err := t.currentlyPlayingController.SeekForward()
		if err != nil {
			log.WithError(err).Error("problem seeking forward")
			return event
		}
	case tcell.KeyRune:
		// attempt to use rune as string
		s := string(event.Rune())
		switch s {
		case " ":
			t.pauseToggle()
		case "=":
			// IDEA: flash the label
			t.currentlyPlayingController.VolumeUp()
		case "-":
			t.currentlyPlayingController.VolumeDown()
		case "+":
			t.currentlyPlayingController.SpeedUp()
		case "_":
			t.currentlyPlayingController.SpeedDown()
		case "]":
			t.skipForward(1)
		case "[":
			t.skipForward(-1)
		case "?":
			log.Debug("switching to help page")
			pages.SwitchToPage("help")
		case "q":
			log.Info("exiting")
			app.Stop()
		}
	}
	return event
}

func (t *TrackPage) pauseToggle() {
	if t.currentlyPlayingController == nil {
		log.Debug("cannot pause, nothing currently playing")
		return
	}

	log.Debug("pausing currently playing track")
	paused := t.currentlyPlayingController.PauseToggle()

	if t.currentlyPlayingRow == 0 {
		log.Debug("nothing currently playing, done toggling pause")
		return
	}

	if paused {
		t.setTrackRowStyle(t.currentlyPlayingRow, t.theme.SecondaryTextColor, trackIconPausedText)
	} else {
		t.setTrackRowStyle(t.currentlyPlayingRow, t.theme.TertiaryTextColor, trackIconPlayingText)
	}
}

// cellConfirmed is called when a user presses enter on a selected cell.
func (t *TrackPage) cellChosen(row, column int) {
	// clear any lingering log messages.
	// TODO: maybe fire this off at an interval later
	t.logBox.Clear()

	log.Debugf("selecting row %d column %d", row, column)

	if row > len(t.tracks) {
		log.Warnf("row out of range %d column %d, length %d", row, column, len(t.tracks))
		return
	}

	track := t.tracks[row-1]

	if t.currentlyPlayingRow != 0 && t.currentlyPlayingController != nil && t.currentlyPlayingTrack != nil {
		log.WithFields(log.Fields{
			"title": t.currentlyPlayingTrack.Title,
		}).Debug("stopping currently playing track")
		t.stopCurrentlyPlaying()
	}

	t.currentlyPlayingRow = row

	// set currently playing row style
	t.setTrackRowStyle(t.currentlyPlayingRow, t.theme.TertiaryTextColor, trackIconPlayingText)

	t.playTrack(&track)
}

// setTrackRowStyle sets the style of a track row. Used for selection, pausing,
// unpausing, etc.
func (t *TrackPage) setTrackRowStyle(row int, color tcell.Color, statusColumnText string) {
	t.trackBox.GetCell(row, columnStatus).SetText(statusColumnText)
	t.trackBox.GetCell(row, columnArtist).SetTextColor(color)
	t.trackBox.GetCell(row, columnAlbum).SetTextColor(color)
	t.trackBox.GetCell(row, columnTrack).SetTextColor(color)
}

func (t *TrackPage) playTrack(track *library.Track) {
	log.WithFields(log.Fields{
		"name": track.Title,
		"path": track.Path,
	}).Debug("playing track")

	controller, err := t.player.Play(*track, false)
	if err != nil {
		log.WithError(err).Fatal("could not play file")
		return
	}

	t.currentlyPlayingController = controller
	t.currentlyPlayingTrack = track
}

// audioPlaying is a loop that checks on currently playing track
// progress
func (t *TrackPage) audioPlaying(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("context done")
			return
		default:
			t.checkCurrentlyPlaying()
			time.Sleep(checkAudioMillis * time.Millisecond)
		}
	}
}

func (t *TrackPage) stopCurrentlyPlaying() {
	if t.currentlyPlayingController == nil {
		return
	}

	t.setTrackRowStyle(t.currentlyPlayingRow, t.theme.PrimaryTextColor, trackIconEmptyText)

	t.currentlyPlayingController.Stop()
	t.currentlyPlayingController = nil
	t.currentlyPlayingTrack = nil
	t.currentlyPlayingRow = 0
}

// if audio is playing, update status, if stopped, clear
func (t *TrackPage) checkCurrentlyPlaying() {
	if t.currentlyPlayingController == nil || t.currentlyPlayingTrack == nil {
		return
	}

	prog, err := t.currentlyPlayingController.Progress()
	if err != nil {
		log.WithError(err).Error("could not get audio progress")
	}

	t.updateProgress(prog, t.currentlyPlayingTrack)

	// check if audio has stopped
	if prog.Finished {
		log.Debug("track has finished playing")

		// move to next track
		t.skipForward(1)
	}
}

// skipForward skips forward/backward on the playlist. count can be negative to go backward.
//
// TODO: add unit tests for next track logic
func (t *TrackPage) skipForward(count int) {
	// attempt to play the next track available
	nextRow := t.currentlyPlayingRow + count

	// if skipping too far ahead, go to beginning
	if nextRow <= 0 {
		nextRow = len(t.tracks)
	}

	// if we're at the end of the list, start over
	if t.currentlyPlayingRow >= len(t.tracks) && count > 0 {
		nextRow = 1
	}

	log.WithFields(log.Fields{
		"nextRow":             nextRow,
		"currentlyPlayingRow": t.currentlyPlayingRow,
		"totalTracks":         len(t.tracks),
	}).Debug("playing next track")

	t.cellChosen(nextRow, columnStatus)
}

func (t *TrackPage) updateProgress(prog player.PlayState, track *library.Track) {
	percentageComplete := int(prog.Progress * 100)

	log.WithFields(log.Fields{
		"progress":   prog.Progress,
		"position":   prog.Position,
		"volume":     prog.Volume,
		"speed":      prog.Speed,
		"track":      track.Title,
		"goroutines": runtime.NumGoroutine(),
	}).Trace("progress")

	app.QueueUpdateDraw(func() {
		t.progressBox.SetCell(0, 0, tview.NewTableCell("Title"))
		t.progressBox.SetCell(0, 1, &tview.TableCell{Text: track.Title, Color: t.theme.TertiaryTextColor})
		t.progressBox.SetCell(1, 0, tview.NewTableCell("Album"))
		t.progressBox.SetCell(1, 1, &tview.TableCell{Text: track.Album, Color: t.theme.TertiaryTextColor})
		t.progressBox.SetCell(2, 0, tview.NewTableCell("Artist"))
		t.progressBox.SetCell(2, 1, &tview.TableCell{Text: track.Artist, Color: t.theme.TertiaryTextColor})

		t.progressBox.SetCell(0, 2, tview.NewTableCell("Progress"))
		t.progressBox.SetCell(0, 3, &tview.TableCell{Text: fmt.Sprintf("%s %d%%", prog.Position, percentageComplete), Color: t.theme.TertiaryTextColor})
		t.progressBox.SetCell(1, 2, &tview.TableCell{Text: "Volume"})
		t.progressBox.SetCell(1, 2, tview.NewTableCell("Volume"))
		t.progressBox.SetCell(1, 3, &tview.TableCell{Text: prog.Volume, Color: t.theme.TertiaryTextColor})
		t.progressBox.SetCell(2, 2, tview.NewTableCell("Speed"))
		t.progressBox.SetCell(2, 3, &tview.TableCell{Text: prog.Speed, Color: t.theme.TertiaryTextColor})

		t.progressBox.SetCell(3, 0, tview.NewTableCell("Path"))
		t.progressBox.SetCell(3, 1, &tview.TableCell{Text: track.Path, Color: t.theme.TertiaryTextColor})

	})
}

func (t *TrackPage) welcome() {
	t.progressBox.Clear().
		SetCell(0, 0, tview.NewTableCell("grump")).
		SetCell(0, 1, &tview.TableCell{Text: fmt.Sprintf("%s", build.Version), Color: t.theme.TitleColor, NotSelectable: true}).
		SetCell(1, 0, tview.NewTableCell("files scanned")).
		SetCell(1, 1, &tview.TableCell{Text: fmt.Sprintf("%d", len(t.tracks)), Color: t.theme.SecondaryTextColor, NotSelectable: true}).
		SetCell(2, 0, tview.NewTableCell("for help, press")).
		SetCell(2, 1, &tview.TableCell{Text: "?", Color: t.theme.TertiaryTextColor, NotSelectable: true})
	//SetCell(3, 0, tview.NewTableCell("report bugs at")).
	//SetCell(3, 1, &tview.TableCell{Text: "github.com/dhulihan/grump", Color: t.theme.GraphicsColor, NotSelectable: true})
}

func (t *TrackPage) trackColumns(table *tview.Table) {
	table.
		SetCell(0, columnStatus, &tview.TableCell{Text: trackIconEmptyText, Color: t.theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnArtist, &tview.TableCell{Text: "Artist", Color: t.theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnAlbum, &tview.TableCell{Text: "Album", Color: t.theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnTrack, &tview.TableCell{Text: "Track", Color: t.theme.TitleColor, NotSelectable: true})
}

func (t *TrackPage) trackCell(table *tview.Table, row int, track library.Track) {
	table.
		SetCell(row, columnStatus, &tview.TableCell{Text: trackIconEmptyText, Color: t.theme.PrimaryTextColor}).
		SetCell(row, columnArtist, &tview.TableCell{Text: track.Artist, Color: t.theme.PrimaryTextColor, Expansion: 4, MaxWidth: 8}).
		SetCell(row, columnAlbum, &tview.TableCell{Text: track.Album, Color: t.theme.PrimaryTextColor, Expansion: 4, MaxWidth: 8}).
		SetCell(row, columnTrack, &tview.TableCell{Text: track.Title, Color: t.theme.PrimaryTextColor, Expansion: 10, MaxWidth: 8})
}
