package ui

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/dhulihan/ghoulish/library"
	"github.com/dhulihan/ghoulish/player"
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
	checkAudioMillis = 250
)

var (
	app         *tview.Application // The tview application.
	pages       *tview.Pages       // The application pages.
	finderFocus tview.Primitive    // The primitive in the Finder that last had focus.
)

// Start starts the ui
func Start(ctx context.Context, db *library.Library, musicPlayer player.AudioPlayer) error {
	// hard code first for now
	musicLibrary := db.AudioShelves[0]
	app = tview.NewApplication()
	start(ctx, musicLibrary, musicPlayer)
	if err := app.Run(); err != nil {
		return fmt.Errorf("Error running application: %s", err)
	}

	return nil
}

// Sets up a "Finder" used to navigate the artists, albums, and tracks.
func start(ctx context.Context, ml library.AudioShelf, pl player.AudioPlayer) {
	// Set up the pages
	trackPage := NewTrackPage(ctx, ml, pl)

	pages = tview.NewPages().
		AddPage("home", trackPage.Page(ctx), true, true)
	app.SetRoot(pages, true)
}

// TrackPage is a page that displays playable audio tracks
type TrackPage struct {
	tracks                     []library.Track
	player                     player.AudioPlayer
	currentlyPlayingController *player.Controller
	currentlyPlayingTrack      *library.Track
	currentlyPlayingRow        int

	// layout
	left   *tview.List
	center *tview.Flex
	top    *tview.Box
	middle *tview.Table
	bottom *tview.Table

	theme *tview.Theme
}

func defaultTheme() *tview.Theme {
	return &tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorBlack,    // Main background color for primitives.
		ContrastBackgroundColor:     tcell.ColorBlue,     // Background color for contrasting elements.
		MoreContrastBackgroundColor: tcell.ColorGreen,    // Background color for even more contrasting elements.
		BorderColor:                 tcell.ColorWhite,    // Box borders.
		TitleColor:                  tcell.ColorYellow,   // Box titles.
		GraphicsColor:               tcell.ColorWhite,    // Graphics.
		PrimaryTextColor:            tcell.ColorWhite,    // Primary text.
		SecondaryTextColor:          tcell.ColorRed,      // Secondary text (e.g. labels).
		TertiaryTextColor:           tcell.ColorGreen,    // Tertiary text (e.g. subtitles, notes).
		InverseTextColor:            tcell.ColorBlue,     // Text on primary-colored backgrounds.
		ContrastSecondaryTextColor:  tcell.ColorDarkCyan, // Secondary text on ContrastBackgroundColor-colored backgrounds.
	}
}

// NewTrackPage generates the track page
func NewTrackPage(ctx context.Context, ml library.AudioShelf, pl player.AudioPlayer) *TrackPage {
	// Create the basic objects.
	middle := tview.NewTable().SetBorders(true)

	top := tview.NewBox().SetBorder(true)
	bottom := tview.NewTable()
	bottom.SetBorder(true)

	return &TrackPage{
		tracks: ml.Tracks(),
		player: pl,
		top:    top,
		middle: middle,
		bottom: bottom,
		theme:  defaultTheme(),
	}
}

// Page populates the layout for the track page
func (t *TrackPage) Page(ctx context.Context) tview.Primitive {
	t.trackColumns(t.middle)

	for i, track := range t.tracks {
		// incr by one to pass table headers
		t.trackCell(t.middle, i+1, track)
	}

	t.middle.
		// fired on Escape, Tab, or Backtab key
		SetDoneFunc(func(key tcell.Key) {
			log.Debugf("done func firing, key [%v]", key)
		}).
		SetSelectable(true, false).SetSelectedFunc(t.cellChosen).SetSelectedStyle(t.theme.SecondaryTextColor, t.theme.PrimitiveBackgroundColor, tcell.AttrNone)
	//t.middle.SetSelectedStyle(t.theme.TertiaryTextColor, t.theme.PrimitiveBackgroundColor, tcell.AttrNone)

	t.middle.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// placeholder nil check for convenience

		log.Debugf("input capture firing, name [%s] key [%d] rune [%s]", event.Name(), event.Key(), string(event.Rune()))
		switch event.Key() {
		case tcell.KeyLeft:
			// return if nothing is currently playing
			// TODO: find a better way to perform this check
			if t.currentlyPlayingController == nil {
				return event
			}

			err := t.currentlyPlayingController.SeekBackward()
			if err != nil {
				log.WithError(err).Error("problem skipping forward")
				return event
			}
		case tcell.KeyRight:
			// return if nothing is currently playing
			if t.currentlyPlayingController == nil {
				return event
			}

			err := t.currentlyPlayingController.SeekForward()
			if err != nil {
				log.WithError(err).Error("problem skipping forward")
				return event
			}

		// KeyRune can be any rune
		case tcell.KeyRune:
			// return if nothing is currently playing
			if t.currentlyPlayingController == nil {
				return event
			}

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
			}
		}
		return event
	})

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.top, 0, 1, false).
		AddItem(t.middle, 0, 3, true).
		AddItem(t.bottom, 6, 1, false)

	// Create the layout.
	flex := tview.NewFlex().
		AddItem(main, 0, 3, true)

	// one outstanding goroutine that tracks audio progress
	go t.audioPlaying(ctx)

	return flex
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
	log.Debugf("selecting row %d column %d", row, column)

	if row > len(t.tracks) {
		log.Warnf("row out of range %d column %d, length %d", row, column, len(t.tracks))
		return
	}

	track := t.tracks[row-1]

	if t.currentlyPlayingRow != 0 {
		// clear previous row style
		t.setTrackRowStyle(t.currentlyPlayingRow, t.theme.PrimaryTextColor, trackIconEmptyText)
	}

	t.currentlyPlayingRow = row

	// set currently playing row style
	t.setTrackRowStyle(t.currentlyPlayingRow, t.theme.TertiaryTextColor, trackIconPlayingText)

	t.playTrack(&track)
}

// setTrackRowStyle sets the style of a track row. Used for selection, pausing,
// unpausing, etc.
func (t *TrackPage) setTrackRowStyle(row int, color tcell.Color, statusColumnText string) {
	t.middle.GetCell(row, columnStatus).SetText(statusColumnText)
	t.middle.GetCell(row, columnArtist).SetTextColor(color)
	t.middle.GetCell(row, columnAlbum).SetTextColor(color)
	t.middle.GetCell(row, columnTrack).SetTextColor(color)
}

func (t *TrackPage) playTrack(track *library.Track) {
	if t.currentlyPlayingController != nil && t.currentlyPlayingTrack != nil {
		log.WithFields(log.Fields{
			"title": t.currentlyPlayingTrack.Title,
		}).Debug("stopping currently playing track")
		t.stopCurrentlyPlaying()
	}

	log.WithFields(log.Fields{
		"name": track.Title,
		"path": track.Path,
	}).Info("playing track")

	controller, err := t.player.Play(*track, false)
	if err != nil {
		log.WithError(err).Fatal("could not play file")
		return
	}

	t.currentlyPlayingController = controller
	t.currentlyPlayingTrack = track
}

// TODO add directions to chan
func (t *TrackPage) audioPlaying(ctx context.Context) {
	log.Debug("audio playing")

	for {
		select {
		case <-ctx.Done():
			log.Debug("context done")
			return
		default:
			// since this is running inside a goroutine, we must safely perform
			// changes to the tview application using QueueUpdate to avoid race
			// conditions.
			t.checkCurrentlyPlaying()
			time.Sleep(checkAudioMillis * time.Millisecond)
		}
	}
}

func (t *TrackPage) stopCurrentlyPlaying() {
	// TODO: lock this
	t.currentlyPlayingController.Stop()
	t.currentlyPlayingController = nil
	t.currentlyPlayingTrack = nil
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

		// TODO: lock since this is accessed from multiple threads
		t.currentlyPlayingController = nil
		t.currentlyPlayingTrack = nil

		if t.currentlyPlayingRow == 0 {
			return
		}

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
	}).Debug("progress")

	app.QueueUpdateDraw(func() {
		t.bottom.SetCell(0, 0, tview.NewTableCell("Title"))
		t.bottom.SetCell(0, 1, &tview.TableCell{Text: track.Title, Color: t.theme.TertiaryTextColor})
		t.bottom.SetCell(1, 0, tview.NewTableCell("Album"))
		t.bottom.SetCell(1, 1, &tview.TableCell{Text: track.Album, Color: t.theme.TertiaryTextColor})
		t.bottom.SetCell(2, 0, tview.NewTableCell("Artist"))
		t.bottom.SetCell(2, 1, &tview.TableCell{Text: track.Artist, Color: t.theme.TertiaryTextColor})

		t.bottom.SetCell(0, 2, tview.NewTableCell("Progress"))
		t.bottom.SetCell(0, 3, &tview.TableCell{Text: fmt.Sprintf("%s %d%%", prog.Position, percentageComplete), Color: t.theme.TertiaryTextColor})
		t.bottom.SetCell(1, 2, &tview.TableCell{Text: "Volume"})
		t.bottom.SetCell(1, 2, tview.NewTableCell("Volume"))
		t.bottom.SetCell(1, 3, &tview.TableCell{Text: prog.Volume, Color: t.theme.TertiaryTextColor})
		t.bottom.SetCell(2, 2, tview.NewTableCell("Speed"))
		t.bottom.SetCell(2, 3, &tview.TableCell{Text: prog.Speed, Color: t.theme.TertiaryTextColor})

		t.bottom.SetCell(3, 0, tview.NewTableCell("Path"))
		t.bottom.SetCell(3, 1, &tview.TableCell{Text: track.Path, Color: t.theme.TertiaryTextColor})

	})
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
