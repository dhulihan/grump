package ui

import (
	"context"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type LogsPage struct {
	theme *tview.Theme
}

func NewLogsPage(ctx context.Context) *LogsPage {
	theme := defaultTheme()

	return &LogsPage{
		theme: theme,
	}
}

// Page populates the layout for the help page
func (p *LogsPage) Page(ctx context.Context) tview.Primitive {
	logs.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		globalInputCapture(event)

		switch event.Key() {
		case tcell.KeyESC:
			pages.SwitchToPage("tracks")
		}

		return event
	})

	bottom := tview.NewTextView().SetText("Press escape to go back.")

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		//AddItem(p.middle, 0, 6, true).
		AddItem(logs, 0, 6, true).
		AddItem(bottom, 1, 0, false)

	// Create the layout.
	flex := tview.NewFlex().
		AddItem(main, 0, 3, true)

	return flex
}
