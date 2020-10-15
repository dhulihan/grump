package ui

import (
	"github.com/gdamore/tcell"
)

const (
	ratingMin = 0.0
	ratingMax = 255
	scoreMin  = 0.0
	scoreMax  = 5.0

	Score00 = "🌑"
	Score05 = "🌗"
	Score10 = "🌕"
	Score15 = "🌕🌗"
	Score20 = "🌕🌕"
	Score25 = "🌕🌕🌗"
	Score30 = "🌕🌕🌕"
	Score35 = "🌕🌕🌕🌗"
	Score40 = "🌕🌕🌕🌕"
	Score45 = "🌕🌕🌕🌕🌗"
	Score50 = "🌕🌕🌕🌕🌕"

	// mapping of track uint8 ratings to human friendly scores
	Rating00 = 0
	Rating05 = 13
	Rating10 = 1
	Rating15 = 54
	Rating20 = 64
	Rating25 = 118
	Rating30 = 128
	Rating35 = 186
	Rating40 = 196
	Rating45 = 242
	Rating50 = 255
)

var (
	Scores = []string{
		Score00,
		Score05,
		Score10,
		Score15,
		Score20,
		Score25,
		Score30,
		Score35,
		Score40,
		Score45,
		Score50,
	}
)

// Score returns a human-friendly rating string and color. It clamsp 0-255 to a
// 0-5 rating string (think 5 stars).
func Score(rating uint8) string {
	switch {
	case rating == Rating00:
		return Score00
	case rating <= Rating10:
		return Score10
	// yes this is bizarre
	case rating <= Rating05:
		return Score05
	case rating <= Rating15:
		return Score15
	case rating <= Rating20:
		return Score20
	case rating <= Rating25:
		return Score25
	case rating <= Rating30:
		return Score30
	case rating <= Rating35:
		return Score35
	case rating <= Rating40:
		return Score40
	case rating <= Rating45:
		return Score45
	case rating <= ratingMax:
		return Score50
	default:
		return Score00
	}
}

// ScoreColor returns a color for the score
func ScoreColor(score string) tcell.Color {
	switch score {
	case "0.0":
		return tcell.ColorGrey
	case "0.5":
		return tcell.ColorRed
	case "1.0":
		return tcell.ColorRed
	case "1.5":
		return tcell.ColorRed
	case "2.0":
		return tcell.ColorOrange
	case "2.5":
		return tcell.ColorOrange
	case "3.0":
		return tcell.ColorYellow
	case "3.5":
		return tcell.ColorYellow
	case "4.0":
		return tcell.ColorGreen
	case "4.5":
		return tcell.ColorGreen
	case "5.0":
		return tcell.ColorAqua
	default:
		return tcell.ColorWhite
	}
}

//// Rating converts a human-friendly score value to a rating
//func Rating(score string) (uint8, error) {
//s, err := strconv.ParseFloat(score, 32)
//if err != nil {
//return 0, err
//}

//percentMax := s / scoreMax
//r := percentMax * ratingMax
//return uint8(r), nil
//}

// Rating converts a human-friendly score value to a rating
func Rating(score string) uint8 {
	switch score {
	case Score00:
		return Rating00
	case Score05:
		return Rating05
	case Score10:
		return Rating10
	case Score15:
		return Rating15
	case Score20:
		return Rating20
	case Score25:
		return Rating25
	case Score30:
		return Rating30
	case Score35:
		return Rating35
	case Score40:
		return Rating40
	case Score45:
		return Rating45
	case Score50:
		return Rating50
	default:
		return Rating00
	}
}
