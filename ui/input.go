package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

func modalWrapper(p tview.Primitive, width, height int) *tview.Flex {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, false).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}

func newInputField(label string, text string, done func(key tcell.Key)) *tview.InputField {
	return tview.NewInputField().
		SetFieldWidth(40).
		SetLabel(label).
		SetText(text).
		SetDoneFunc(done)
}

func newDropDown(label string, options []string, current int) *tview.DropDown {
	return tview.NewDropDown().
		SetLabel(label).
		SetOptions(options, nil).
		SetCurrentOption(current)
}

// get the text of an input field belonging to a form
func getFormInputText(f *tview.Form, label string) string {
	i := inputField(f, label)
	if i == nil {
		return ""
	}

	return i.GetText()
}

// input helpers
func inputField(f *tview.Form, label string) *tview.InputField {
	fi := f.GetFormItemByLabel(label)
	if fi == nil {
		log.WithField("label", label).Error("could not find form item")
		return nil
	}

	var input *tview.InputField
	var ok bool
	if input, ok = fi.(*tview.InputField); !ok {
		log.WithField("label", label).Error("could not cast FormItem into InputField")
		return nil
	}

	return input
}

func (t *TrackPage) dropDown(label string) *tview.DropDown {
	fi := editForm.GetFormItemByLabel(label)

	if fi == nil {
		log.WithField("label", label).Error("could not find form item")
	}

	var input *tview.DropDown
	var ok bool
	if input, ok = fi.(*tview.DropDown); !ok {
		log.WithField("label", label).Error("could not cast FormItem into DropDown")
		return nil
	}

	return input
}

// get index of first matching string in slice
func indexOf(s []string, x string) int {
	for i, y := range s {
		if x == y {
			return i
		}
	}

	return -1
}
