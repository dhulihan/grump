package ui

import (
	"context"
	"fmt"
	"io"

	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

var (
	app         *tview.Application // The tview application.
	pages       *tview.Pages       // The application pages.
	finderFocus tview.Primitive    // The primitive in the Finder that last had focus.
	build       BuildInfo
	logs        *tview.TextView
	statusBar   *tview.TextView
	deleteModal *tview.Modal
	editForm    *tview.Form
	editPage    *tview.Flex
	theme       *tview.Theme
)

// BuildInfo contains build-time data for displaying version, etc.
type BuildInfo struct {
	Version string
	Commit  string
}

// Start starts the ui
func Start(ctx context.Context, b BuildInfo, db *library.Library, musicPlayer player.AudioPlayer, loggers []io.Writer) error {
	// hard code first for now
	musicLibrary := db.AudioShelves[0]
	app = tview.NewApplication()
	build = b
	start(ctx, musicLibrary, musicPlayer, loggers)
	if err := app.Run(); err != nil {
		return fmt.Errorf("Error running application: %s", err)
	}

	return nil
}

// start the ui
func start(ctx context.Context, ml library.AudioShelf, pl player.AudioPlayer, loggers []io.Writer) {
	theme = defaultTheme()
	setupLoggers(loggers)

	// Set up the pages
	trackPage := NewTrackPage(ctx, ml, pl)
	helpPage := NewHelpPage(ctx)
	logsPage := NewLogsPage(ctx)

	editForm = tview.NewForm()
	editPage = modalWrapper(editForm, 60, 20)

	deleteModal = tview.NewModal()

	pages = tview.NewPages().
		AddPage("help", helpPage.Page(ctx), true, false).
		AddPage("logs", logsPage.Page(ctx), true, false).
		AddPage("tracks", trackPage.Page(ctx), true, true).
		AddPage("edit", editPage, true, false)

	app.SetRoot(pages, true).SetFocus(trackPage.trackList)
}

func defaultTheme() *tview.Theme {
	return &tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorBlack,          // Main background color for primitives.
		ContrastBackgroundColor:     tcell.ColorBlue,           // Background color for contrasting elements.
		MoreContrastBackgroundColor: tcell.ColorGreen,          // Background color for even more contrasting elements.
		BorderColor:                 tcell.ColorGrey,           // Box borders.
		TitleColor:                  tcell.ColorCoral,          // Box titles.
		GraphicsColor:               tcell.ColorFuchsia,        // Graphics.
		PrimaryTextColor:            tcell.ColorWhite,          // Primary text.
		SecondaryTextColor:          tcell.ColorAqua,           // Secondary text (e.g. labels).
		TertiaryTextColor:           tcell.ColorMediumSeaGreen, // Tertiary text (e.g. subtitles, notes).
		InverseTextColor:            tcell.ColorBlue,           // Text on primary-colored backgrounds.
		ContrastSecondaryTextColor:  tcell.ColorDarkCyan,       // Secondary text on ContrastBackgroundColor-colored backgrounds.
	}
}

// globalInputCapture handles input and behavior that is the same across the
// entire application
var globalInputCapture = func(event *tcell.EventKey) *tcell.EventKey {
	s := string(event.Rune())
	switch s {
	case "l":
		pages.SwitchToPage("logs")
	case "t":
		pages.SwitchToPage("tracks")
	case "?":
		pages.SwitchToPage("help")
	case "q":
		log.Info("exiting")
		app.Stop()
	}

	return event
}

// setup loggers (status bar, file, logs page)
func setupLoggers(loggers []io.Writer) {
	logs = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)

	loggers = append(loggers, logs)

	statusBar = tview.NewTextView().
		SetTextColor(theme.BorderColor)
	loggers = append(loggers, statusBar)

	// combine our log destinations
	log.Trace("creating logger group")
	l := io.MultiWriter(loggers...)
	log.SetOutput(l)

	formatter := &appLogger{}
	log.SetFormatter(formatter)
}

// appLogger is a log formatter for the application
type appLogger struct{}

// Format is a custom log formatter that allows write logrus entries to the ui
func (l *appLogger) Format(entry *log.Entry) ([]byte, error) {
	// clear out the log box before writing text to it
	statusBar.Clear()

	lf := &log.TextFormatter{}
	return lf.Format(entry)
}
