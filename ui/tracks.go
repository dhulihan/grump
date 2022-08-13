package ui

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type trackTarget int

const (
	columnStatus = iota
	columnArtist
	columnAlbum
	columnTrack
	columnRating

	// check audio progess at this interval
	checkAudioMillis = 500

	// track target types
	playing trackTarget = iota
	hovered
	next
	prev
)

// TrackPage is a page that displays playable audio tracks
type TrackPage struct {
	// TODO: extract this to a something ui-agnostic
	shelf                      library.AudioShelf
	tracks                     []library.Track
	player                     player.AudioPlayer
	currentlyPlayingController player.AudioController
	currentlyPlayingTrack      *library.Track
	currentlyPlayingRow        int
	shuffle                    bool

	// layout
	left         *tview.List
	center       *tview.Flex
	logBox       *tview.TextView
	trackList    *tview.Table
	playStateBox *tview.Table
	statusBox    *tview.Table
	editForm     *tview.Form
}

// NewTrackPage generates the track page
func NewTrackPage(ctx context.Context, shelf library.AudioShelf, pl player.AudioPlayer) *TrackPage {

	// Create the basic objects.
	trackList := tview.NewTable().SetBorders(true).SetBordersColor(theme.BorderColor)

	playStateBox := tview.NewTable()
	playStateBox.SetBorder(true).SetBorderColor(theme.BorderColor)

	p := &TrackPage{
		//editForm:    form,
		shelf:        shelf,
		tracks:       shelf.Tracks(),
		player:       pl,
		logBox:       statusBar,
		trackList:    trackList,
		playStateBox: playStateBox,
		statusBox:    tview.NewTable(),
	}

	return p
}

// Page populates the layout for the track page
func (t *TrackPage) Page(ctx context.Context) tview.Primitive {
	t.trackColumns(t.trackList)

	for i, track := range t.tracks {
		// incr by one to pass table headers
		t.trackCell(t.trackList, i+1, track)
	}

	t.trackList.
		// fired on Escape, Tab, or Backtab key
		SetDoneFunc(func(key tcell.Key) {
			log.Debugf("done func firing, key [%v]", key)
		}).
		SetSelectable(true, false).SetSelectedFunc(t.cellChosen).SetSelectedStyle(theme.SecondaryTextColor, theme.PrimitiveBackgroundColor, tcell.AttrNone)
	//t.trackList.SetSelectedStyle(theme.TertiaryTextColor, theme.PrimitiveBackgroundColor, tcell.AttrNone)

	t.trackList.SetInputCapture(t.inputCapture)

	editForm.SetCancelFunc(t.editCancel)

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.trackList, 0, 14, true).
		AddItem(t.playStateBox, 0, 2, false).
		AddItem(t.statusBox, 1, 1, false).
		AddItem(t.logBox, 1, 1, false)

	// Create the layout.
	flex := tview.NewFlex().
		AddItem(main, 0, 3, true)

	t.welcome()

	// one outstanding goroutine that tracks audio progress
	go t.audioPlaying(ctx)

	return flex
}

// main key input handler for this page
func (t *TrackPage) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	// placeholder nil check for convenience
	log.Tracef("input capture firing, name [%s] key [%d] rune [%s]", event.Name(), event.Key(), string(event.Rune()))

	globalInputCapture(event)

	switch event.Key() {
	case tcell.KeyRune:
		// attempt to use rune as string
		s := string(event.Rune())
		switch s {
		case "D":
			t.describe(hovered)
		}
	}

	// something is currently playing, handle that
	if t.currentlyPlayingController != nil {
		return t.currentlyPlayingInputCapture(event)
	}

	return event
}

// track fetches a track
func (t *TrackPage) track(target trackTarget) (*library.Track, error) {
	var track *library.Track

	switch target {
	case playing:
		if t.currentlyPlayingTrack == nil {
			return nil, fmt.Errorf("no track currently playing")
		}

		track = t.currentlyPlayingTrack
	// TODO: hovered does not work yet
	case hovered:
		row, column := t.trackList.GetOffset()
		track = &t.tracks[row]
		log.WithFields(log.Fields{"row": row, "column": column}).Debug("currently hovered track")

		return track, nil
	default:
		return nil, fmt.Errorf("trackTarget not supported: %v", target)
	}

	log.WithFields(log.Fields{"track": track}).Debug("track targeted")
	return track, nil
}

func (t *TrackPage) describe(target trackTarget) {
	track, err := t.track(target)
	if err != nil {
		log.WithError(err).Error("could not target track")
		return
	}

	log.WithFields(log.Fields{
		"title":       track.Title,
		"album":       track.Album,
		"artist":      track.Artist,
		"rating":      track.Rating,
		"ratingEmail": track.RatingEmail,
		"score":       Score(track.Rating),
		"playCount":   track.PlayCount,
	}).Info("describing track")
}

// inputDone is used to enhance to form input movement
func (t *TrackPage) inputDone(key tcell.Key) {
	log.Tracef("modal input capture firing, key [%d] %s", key, tcell.KeyNames[key])
	// perform this asynchronously to avoid weird focus state where the
	// InputField holds on to focus
	go func() {
		app.QueueUpdateDraw(func() {
			switch key {
			case tcell.KeyEnter:
				t.save()
				pages.HidePage("edit")
				editForm.Blur()
			case tcell.KeyEscape:
				pages.HidePage("edit")
				editForm.Blur()
			case tcell.KeyUp:
				fi, _ := editForm.GetFocusedItemIndex()
				index := fi - 1
				editForm.SetFocus(index)
				app.SetFocus(editForm)
			case tcell.KeyDown:
				fi, _ := editForm.GetFocusedItemIndex()
				index := fi + 1
				editForm.SetFocus(index)
				app.SetFocus(editForm)
			}
		})
	}()
}

func (t *TrackPage) editCancel() {
	pages.SwitchToPage("tracks")

	// unpause
	if t.currentlyPlayingController.Paused() {
		t.pauseToggle()
	}
}

func (t *TrackPage) edit(target trackTarget) {
	log.Debug("editing track")

	track, err := t.track(target)
	if err != nil {
		log.WithError(err).Error("could not target track")
		return
	}

	// pause
	if !t.currentlyPlayingController.Paused() {
		t.pauseToggle()
	}

	// if blank, use previous album/artist
	if track.Album == "" {
		track.Album = getFormInputText(editForm, "Album")
	}

	if track.Artist == "" {
		track.Artist = getFormInputText(editForm, "Artist")
	}

	editForm.Clear(true).
		AddFormItem(newInputField("Title", track.Title, t.inputDone)).
		AddFormItem(newInputField("Album", track.Album, t.inputDone)).
		AddFormItem(newInputField("Artist", track.Artist, t.inputDone)).
		AddFormItem(newDropDown("Score", Scores, indexOf(Scores, Score(track.Rating)))).
		AddButton("Save", t.save).
		AddButton("Cancel", t.editCancel)

	editForm.SetBorder(true).SetTitle("Edit Track").SetTitleAlign(tview.AlignLeft)
	pages.ShowPage("edit")
	app.SetFocus(editForm)
}

func (t *TrackPage) save() {
	log.Debug("saving track")
	ctx := context.Background()

	prev, err := t.track(playing)
	if err != nil {
		log.WithError(err).Error("could not target track")
		return
	}
	row := t.currentlyPlayingRow
	track := prev

	_, score := t.dropDown("Score").GetCurrentOption()

	track.Title = inputField(editForm, "Title").GetText()
	track.Album = inputField(editForm, "Album").GetText()
	track.Artist = inputField(editForm, "Artist").GetText()
	track.Rating = Rating(score)

	log.WithFields(log.Fields{
		"title":  track.Title,
		"album":  track.Album,
		"artist": track.Artist,
		"rating": track.Rating,
		"row":    row,
	}).Debug("collected track data from form")

	_, err = t.shelf.SaveTrack(ctx, prev, track)
	if err != nil {
		log.WithField("track", track).WithError(err).Error("could not save track")
		return
	}

	// update track row
	t.trackCell(t.trackList, row, *track)

	// update cache
	t.tracks[row-1] = *track

	// switch back to tracks page
	pages.SwitchToPage("tracks")

	// unpause
	if t.currentlyPlayingController.Paused() {
		t.pauseToggle()
	}
}

func (t *TrackPage) confirmDelete(ctx context.Context, target trackTarget) {
	track, err := t.track(target)
	if err != nil {
		log.WithError(err).Error("could not target track")
		return
	}

	if !t.currentlyPlayingController.Paused() {
		t.pauseToggle()
	}

	msg := fmt.Sprintf(`
	Delete?

	Title: %s
	Album: %s
	Artist: %s

	%s`,
		track.Title,
		track.Album,
		track.Artist,
		track.Path,
	)

	deleteModal = tview.NewModal().
		SetText(msg).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				err := t.deleteTrack(ctx)

				if err != nil {
					log.WithError(err).Error("could not delete track")
				}
			}
			app.SetRoot(pages, true).SetFocus(t.trackList)
		})

	app.SetRoot(deleteModal, false).SetFocus(deleteModal)
}

func (t *TrackPage) deleteTrack(ctx context.Context) error {
	track, err := t.track(playing)
	if err != nil {
		return err
	}
	log.WithField("track", track).Debug("deleting track")

	row := t.currentlyPlayingRow

	// stop playing
	t.stopCurrentlyPlaying()

	// delete from library
	err = t.shelf.DeleteTrack(ctx, track)
	if err != nil {
		return err
	}

	// delete from cache
	removed := t.removeTrackFromCache(row - 1)
	log.WithFields(log.Fields{
		"trackRemovedFromCache": removed,
		"row":                   row,
	}).Debug("track removed from cache")

	// update ui
	t.trackList.RemoveRow(row)

	// log
	log.WithFields(log.Fields{
		"track": track,
	}).Info("deleted track")

	// play next track
	t.cellChosen(row, 0)

	return nil
}

// remove track from cache and return it
func (t *TrackPage) removeTrackFromCache(i int) library.Track {
	track := t.tracks[i]
	t.tracks = append(t.tracks[:i], t.tracks[i+1:]...)
	return track

}

// handle key input while a track is playing
func (t *TrackPage) currentlyPlayingInputCapture(event *tcell.EventKey) *tcell.EventKey {
	ctx := context.Background()

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
	case tcell.KeyDelete:
		t.confirmDelete(ctx, playing)
		return event
	case tcell.KeyRune:
		// attempt to use rune as string
		s := string(event.Rune())
		switch s {
		// key - playing
		// shift+key - hovered
		case "0":
			t.SetScore(Score00)
		case ")":
			t.SetScore(Score05)
		case "1":
			t.SetScore(Score10)
		case "!":
			t.SetScore(Score15)
		case "2":
			t.SetScore(Score20)
		case "@":
			t.SetScore(Score25)
		case "3":
			t.SetScore(Score30)
		case "#":
			t.SetScore(Score35)
		case "4":
			t.SetScore(Score40)
		case "$":
			t.SetScore(Score45)
		case "5":
			t.SetScore(Score50)
		case "d":
			t.describe(playing)
		case "e":
			t.edit(playing)
		case " ":
			t.pauseToggle()
		case "=":
			// IDEA: flash the label
			t.currentlyPlayingController.VolumeUp()
		case "-":
			t.currentlyPlayingController.VolumeDown()
		case "S":
			t.shuffleToggle()
		case "+":
			t.currentlyPlayingController.SpeedUp()
		case "_":
			t.currentlyPlayingController.SpeedDown()
		case "]":
			t.skip(1)
		case "[":
			t.skip(-1)
		case "?":
			log.Trace("switching to help page")
			pages.SwitchToPage("help")
		case "q":
			app.Stop()
		}
	}
	return event
}

func (t *TrackPage) shuffleToggle() {
	// thread safe? nope!
	t.shuffle = !t.shuffle
	log.WithField("enabled", t.shuffle).Debug("toggling shuffle")

	if t.shuffle {
		t.statusBox.SetCell(0, 0, tview.NewTableCell("Shuffle"))
		t.statusBox.SetCell(0, 1, &tview.TableCell{Text: shuffleIconOn, Color: theme.TertiaryTextColor})
	} else {
		t.statusBox.SetCellSimple(0, 0, "")
		t.statusBox.SetCellSimple(0, 1, "")
	}
}

func (t *TrackPage) pauseToggle() {
	if t.currentlyPlayingController == nil {
		log.Debug("cannot pause, nothing currently playing")
		return
	}

	log.Debug("pausing currently playing track")
	t.currentlyPlayingController.PauseToggle()

	if t.currentlyPlayingRow == 0 {
		log.Debug("nothing currently playing, done toggling pause")
		return
	}

	if t.currentlyPlayingController.Paused() {
		t.setTrackRowStyle(t.currentlyPlayingRow, theme.SecondaryTextColor, trackIconPausedText)
	} else {
		t.setTrackRowStyle(t.currentlyPlayingRow, theme.TertiaryTextColor, trackIconPlayingText)
	}
}

// cellConfirmed is called when a user presses enter on a selected cell.
func (t *TrackPage) cellChosen(row, column int) {
	// clear any lingering log messages.
	// TODO: maybe fire this off at an interval later
	t.logBox.Clear()

	if row == 0 {
		log.Info("please select a track")
		return
	}

	log.Tracef("selecting row %d column %d", row, column)

	if row > len(t.tracks) {
		log.Warnf("row out of range %d column %d, length %d", row, column, len(t.tracks))
		return
	}

	track := t.tracks[row-1]

	if t.currentlyPlayingRow != 0 && t.currentlyPlayingController != nil && t.currentlyPlayingTrack != nil {
		log.WithFields(log.Fields{
			"track": t.currentlyPlayingTrack,
			"row":   t.currentlyPlayingRow,
		}).Debug("stopping currently playing track")
		t.stopCurrentlyPlaying()
	}

	t.currentlyPlayingRow = row

	// set currently playing row style
	t.setTrackRowStyle(t.currentlyPlayingRow, theme.TertiaryTextColor, trackIconPlayingText)

	t.playTrack(&track)
}

// setTrackRowStyle sets the style of a track row. Used for selection, pausing,
// unpausing, etc.
func (t *TrackPage) setTrackRowStyle(row int, color tcell.Color, statusColumnText string) {
	t.trackList.GetCell(row, columnStatus).SetText(statusColumnText)
	t.trackList.GetCell(row, columnArtist).SetTextColor(color)
	t.trackList.GetCell(row, columnAlbum).SetTextColor(color)
	t.trackList.GetCell(row, columnTrack).SetTextColor(color)
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
	log.WithField("row", t.currentlyPlayingRow).Debug("clearing track style")
	t.setTrackRowStyle(t.currentlyPlayingRow, theme.PrimaryTextColor, trackIconEmptyText)

	if t.currentlyPlayingController == nil {
		return
	}

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

	ps, err := t.currentlyPlayingController.PlayState()
	if err != nil {
		log.WithError(err).Error("could not get audio play state")
	}

	t.updatePlayState(ps, t.currentlyPlayingTrack)

	// check if audio has stopped
	if ps.Finished {
		log.Debug("track has finished playing")

		// move to next track
		t.skip(1)
	}
}

// skip skips forward/backward on the playlist. count can be negative to go backward.
//
// TODO: add unit tests for next track logic
func (t *TrackPage) skip(count int) {
	// attempt to play the next track available
	nextRow := t.currentlyPlayingRow + count

	// if shuffling, choose one at random
	if t.shuffle {
		nextRow = rand.Intn(len(t.tracks))
	}

	// if skipping too far ahead, go to beginning
	if nextRow <= 0 {
		nextRow = len(t.tracks)
	}

	// if we're at the end of the list, start over
	if t.currentlyPlayingRow >= len(t.tracks) && count > 0 {
		nextRow = 1
	}

	log.WithFields(log.Fields{
		"currentlyPlayingRow": t.currentlyPlayingRow,
		"nextRow":             nextRow,
		"totalTracks":         len(t.tracks),
		"skip":                count,
	}).Debug("skipping to next track")

	t.cellChosen(nextRow, columnStatus)
}

func (t *TrackPage) updatePlayState(ps player.PlayState, track *library.Track) {
	percentageComplete := int(ps.Progress * 100)

	log.WithFields(log.Fields{
		"progress":   ps.Progress,
		"position":   ps.Position,
		"volume":     ps.Volume,
		"speed":      ps.Speed,
		"track":      track.Title,
		"goroutines": runtime.NumGoroutine(),
	}).Trace("play state update")

	app.QueueUpdateDraw(func() {
		t.playStateBox.SetCell(0, 0, tview.NewTableCell("Title"))
		t.playStateBox.SetCell(0, 1, &tview.TableCell{Text: track.Title, Color: theme.TertiaryTextColor})
		t.playStateBox.SetCell(1, 0, tview.NewTableCell("Album"))
		t.playStateBox.SetCell(1, 1, &tview.TableCell{Text: track.Album, Color: theme.TertiaryTextColor})
		t.playStateBox.SetCell(2, 0, tview.NewTableCell("Artist"))
		t.playStateBox.SetCell(2, 1, &tview.TableCell{Text: track.Artist, Color: theme.TertiaryTextColor})

		t.playStateBox.SetCell(0, 2, tview.NewTableCell("Progress"))
		t.playStateBox.SetCell(0, 3, &tview.TableCell{Text: fmt.Sprintf("%s %d%%", ps.Position, percentageComplete), Color: theme.TertiaryTextColor})
		t.playStateBox.SetCell(1, 2, &tview.TableCell{Text: "Volume"})
		t.playStateBox.SetCell(1, 2, tview.NewTableCell("Volume"))
		t.playStateBox.SetCell(1, 3, &tview.TableCell{Text: ps.Volume, Color: theme.TertiaryTextColor})
		t.playStateBox.SetCell(2, 2, tview.NewTableCell("Speed"))
		t.playStateBox.SetCell(2, 3, &tview.TableCell{Text: ps.Speed, Color: theme.TertiaryTextColor})
	})
}

func (t *TrackPage) welcome() {
	t.playStateBox.Clear().
		SetCell(0, 0, tview.NewTableCell("grump")).
		SetCell(0, 1, &tview.TableCell{Text: fmt.Sprintf("%s", build.Version), Color: theme.TitleColor, NotSelectable: true}).
		SetCell(1, 0, tview.NewTableCell("files scanned")).
		SetCell(1, 1, &tview.TableCell{Text: fmt.Sprintf("%d", len(t.tracks)), Color: theme.SecondaryTextColor, NotSelectable: true}).
		SetCell(2, 0, tview.NewTableCell("for help, press")).
		SetCell(2, 1, &tview.TableCell{Text: "?", Color: theme.TertiaryTextColor, NotSelectable: true})
}

func (t *TrackPage) trackColumns(table *tview.Table) {
	table.
		SetCell(0, columnStatus, &tview.TableCell{Text: trackIconEmptyText, Color: theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnArtist, &tview.TableCell{Text: "Artist", Color: theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnAlbum, &tview.TableCell{Text: "Album", Color: theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnTrack, &tview.TableCell{Text: "Title", Color: theme.TitleColor, NotSelectable: true}).
		SetCell(0, columnRating, &tview.TableCell{Text: "Rating", Color: theme.TitleColor, NotSelectable: true})
}

func (t *TrackPage) SetScore(score string) {
	ctx := context.Background()
	log.WithFields(log.Fields{"score": score}).Debug("setting score")

	track := t.currentlyPlayingTrack
	row := t.currentlyPlayingRow

	// convert rating
	rating := Rating(score)
	track.Rating = rating
	_, err := t.shelf.SaveTrack(ctx, nil, track)
	if err != nil {
		log.WithError(err).WithField("rating", rating).Error("could not set rating on track")
		return
	}

	// update track row
	t.trackCell(t.trackList, row, *track)

	// restore "playing" visual state
	t.setTrackRowStyle(t.currentlyPlayingRow, theme.TertiaryTextColor, trackIconPlayingText)

	// update cache
	t.tracks[row-1] = *track
}

func (t *TrackPage) trackCell(table *tview.Table, row int, track library.Track) {
	title := track.Title

	// use path if title is empty
	if track.Title == "" {
		title = track.Path
	}

	scoreText := Score(track.Rating)
	scoreColor := ScoreColor(scoreText)

	table.
		SetCell(row, columnStatus, &tview.TableCell{Text: trackIconEmptyText, Color: theme.PrimaryTextColor}).
		SetCell(row, columnArtist, &tview.TableCell{Text: track.Artist, Color: theme.PrimaryTextColor, Expansion: 4, MaxWidth: 8}).
		SetCell(row, columnAlbum, &tview.TableCell{Text: track.Album, Color: theme.PrimaryTextColor, Expansion: 4, MaxWidth: 8}).
		SetCell(row, columnTrack, &tview.TableCell{Text: title, Color: theme.PrimaryTextColor, Expansion: 10, MaxWidth: 8}).
		SetCell(row, columnRating, &tview.TableCell{Text: scoreText, Color: scoreColor})
}
