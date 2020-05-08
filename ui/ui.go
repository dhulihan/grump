package ui

import (
	"context"
	"fmt"

	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	app         *tview.Application // The tview application.
	pages       *tview.Pages       // The application pages.
	finderFocus tview.Primitive    // The primitive in the Finder that last had focus.
	build       BuildInfo
)

type BuildInfo struct {
	Version string
	Commit  string
}

// Start starts the ui
func Start(ctx context.Context, b BuildInfo, db *library.Library, musicPlayer player.AudioPlayer) error {
	// hard code first for now
	musicLibrary := db.AudioShelves[0]
	app = tview.NewApplication()
	build = b
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
	helpPage := NewHelpPage(ctx)

	pages = tview.NewPages().
		AddPage("help", helpPage.Page(ctx), true, true).
		AddPage("tracks", trackPage.Page(ctx), true, true)
	app.SetRoot(pages, true).SetFocus(trackPage.trackBox)
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
