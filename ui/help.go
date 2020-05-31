package ui

import (
	"context"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type HelpPage struct {
	middle *tview.Table
	bottom *tview.Box
	theme  *tview.Theme
	kb     []KeyboardShortcut
}

func NewHelpPage(ctx context.Context) *HelpPage {
	theme := defaultTheme()

	kb := []KeyboardShortcut{
		KeyboardShortcut{"space", "pause/unpause"},
		KeyboardShortcut{"escape", "stop track"},
		KeyboardShortcut{"d", "describe currently playing track"},
		KeyboardShortcut{"e", "edit currently playing track"},
		KeyboardShortcut{"delete", "delete currently playing track (with prompt)"},
		KeyboardShortcut{"l", "view logs page"},
		KeyboardShortcut{"left", "seek forward (does not work on flac)"},
		KeyboardShortcut{"right", "seek backward  (does not work on flac)"},
		KeyboardShortcut{"]", "play next track"},
		KeyboardShortcut{"[", "play previous track"},
		KeyboardShortcut{"=", "volume up"},
		KeyboardShortcut{"-", "volume down"},
		KeyboardShortcut{"+", "speed up"},
		KeyboardShortcut{"_", "speed down"},
		KeyboardShortcut{"q", "quit"},
		KeyboardShortcut{"0", "set rating of currently playing track to " + Score00},
		KeyboardShortcut{"shift+0", "set rating of currently playing track to " + Score05},
		KeyboardShortcut{"1", "set rating of currently playing track to " + Score10},
		KeyboardShortcut{"shift+1", "set rating of currently playing track to " + Score15},
		KeyboardShortcut{"2", "set rating of currently playing track to " + Score20},
		KeyboardShortcut{"shift+2", "set rating of currently playing track to " + Score25},
		KeyboardShortcut{"3", "set rating of currently playing track to " + Score30},
		KeyboardShortcut{"shift+3", "set rating of currently playing track to " + Score35},
		KeyboardShortcut{"4", "set rating of currently playing track to " + Score40},
		KeyboardShortcut{"shift+4", "set rating of currently playing track to " + Score45},
		KeyboardShortcut{"5", "set rating of currently playing track to " + Score50},
	}

	middle := tview.NewTable().SetBorders(true).SetBordersColor(theme.BorderColor)
	return &HelpPage{
		kb:     kb,
		middle: middle,
		theme:  theme,
	}
}

// Page populates the layout for the help page
func (p *HelpPage) Page(ctx context.Context) tview.Primitive {
	p.middle.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		globalInputCapture(event)

		switch event.Key() {
		case tcell.KeyESC:
			pages.SwitchToPage("tracks")
		}

		return event
	})

	p.setupKeyboardShortcuts()

	bottom := tview.NewTextView().SetText("Press escape to go back.")

	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p.middle, 0, 6, true).
		AddItem(bottom, 1, 0, false)

	// Create the layout.
	flex := tview.NewFlex().
		AddItem(main, 0, 3, true)

	return flex
}

// KeyboardShortcut describes a page-specific keyboard shortcut
type KeyboardShortcut struct {
	Key         string
	Description string
}

func (p *HelpPage) keyboardShortcut(row, column int, key, description string) *tview.TableCell {
	return &tview.TableCell{Text: trackIconEmptyText, Color: p.theme.TitleColor, NotSelectable: true}
}

func (p *HelpPage) setupKeyboardShortcuts() {
	for i, kb := range p.kb {
		p.middle.SetCell(i, 0, tview.NewTableCell(kb.Key)).
			SetCell(i, 1, &tview.TableCell{Text: kb.Description, Color: p.theme.TertiaryTextColor})
	}
}
