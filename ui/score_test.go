package ui_test

import (
	"testing"

	"github.com/dhulihan/grump/ui"
	"github.com/gdamore/tcell"
)

func TestScore(t *testing.T) {
	var tests = []struct {
		score  string
		rating uint8
		err    error
	}{
		{ui.Score00, 0, nil},
		{ui.Score05, 13, nil},
		{ui.Score10, 1, nil},
		{ui.Score15, 54, nil},
		{ui.Score20, 64, nil},
		{ui.Score25, 118, nil},
		{ui.Score30, 128, nil},
		{ui.Score35, 186, nil},
		{ui.Score40, 196, nil},
		{ui.Score45, 242, nil},
		{ui.Score50, 255, nil},
	}

	for _, test := range tests {
		score := ui.Score(test.rating)
		if score != test.score {
			t.Errorf("for %d, wanted %s, got %s", test.rating, test.score, score)
		}
	}
}

func TestScoreScolor(t *testing.T) {
	var tests = []struct {
		score string
		color tcell.Color
	}{
		{"0.0", tcell.ColorGrey},
		{"4.5", tcell.ColorGreen},
		{"5.0", tcell.ColorAqua},
	}

	for _, test := range tests {
		color := ui.ScoreColor(test.score)
		if color != test.color {
			t.Errorf("wanted %#v, got %#v", test.color, color)
		}
	}
}

func TestRating(t *testing.T) {
	var tests = []struct {
		score  string
		rating uint8
		err    error
	}{
		{ui.Score00, 0, nil},
		{ui.Score05, 13, nil},
		{ui.Score10, 1, nil},
		{ui.Score15, 54, nil},
		{ui.Score20, 64, nil},
		{ui.Score25, 118, nil},
		{ui.Score30, 128, nil},
		{ui.Score35, 186, nil},
		{ui.Score40, 196, nil},
		{ui.Score45, 242, nil},
		{ui.Score50, 255, nil},
	}

	for _, test := range tests {
		rating := ui.Rating(test.score)
		if rating != test.rating {
			t.Errorf("wanted %d, got %d", test.rating, rating)
		}
	}
}
