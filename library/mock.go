package library

import "context"

type MockAudioLibrary struct {
	tracks []Track
}

func NewMockAudioLibrary(tracks []Track) AudioShelf {
	l := MockAudioLibrary{
		tracks: tracks,
	}

	return &l
}

// Tracks --
func (l *MockAudioLibrary) Tracks() []Track {
	return l.tracks
}

func (l *MockAudioLibrary) LoadTracks() (uint64, error) {
	return uint64(len(l.tracks)), nil
}

func (l *MockAudioLibrary) LoadTrack(ctx context.Context, location string) (*Track, error) {
	return nil, nil
}

func (l *MockAudioLibrary) SaveTrack(ctx context.Context, prev, track *Track) (*Track, error) {
	return nil, nil
}

func (l *MockAudioLibrary) DeleteTrack(ctx context.Context, track *Track) error {
	return nil
}
