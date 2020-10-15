package ui

import (
	"context"
	"fmt"
	"testing"

	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/stretchr/testify/suite"
)

type TrackPageSuite struct {
	suite.Suite
	page *TrackPage
}

func (s *TrackPageSuite) SetupSuite() {
}

func (s *TrackPageSuite) SetupTest() {
	ctx := context.Background()
	pl := player.NewMockAudioPlayer()

	shelf := library.NewMockAudioLibrary(s.mockTracks())
	s.page = NewTrackPage(ctx, shelf, pl)
}

func (s *TrackPageSuite) mockTracks() []library.Track {
	// generate mocktracks
	tracks := make([]library.Track, 5)

	for i := 1; i <= 5; i++ {
		tracks[i-1] = library.Track{
			Title: fmt.Sprintf("Mock Track %d", i),
			Path:  fmt.Sprintf("mock-track-path-%d", i),
		}

	}
	return tracks
}

func (s *TrackPageSuite) TestDeleteTrack() {
	// play track
	//s.page.playTrack(&s.page.tracks[1])
	s.page.cellChosen(2, 0)

	s.Equal(&s.page.tracks[1], s.page.currentlyPlayingTrack)
	s.Equal(2, s.page.currentlyPlayingRow)

	err := s.page.deleteTrack()
	if s.NoError(err) {
		s.Equal(&s.page.tracks[1], s.page.currentlyPlayingTrack)
		s.Equal(2, s.page.currentlyPlayingRow)
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTrackPageSuite(t *testing.T) {
	suite.Run(t, new(TrackPageSuite))
}
