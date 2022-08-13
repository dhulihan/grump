package library

import (
	"context"
)

// AudioShelf is an abstract collection of audio. A shelf has one source type
// (local, internet, spotify account, etc.). For example, a LocalAudioShelf
// contains files stored in a local filesystem.
type AudioShelf interface {
	Tracks() []Track

	// LoadTracks fills the shelf with tracks
	LoadTracks() (count uint64, err error)
	LoadTrack(ctx context.Context, location string) (*Track, error)
	SaveTrack(ctx context.Context, prev, track *Track) (*Track, error)
	DeleteTrack(ctx context.Context, track *Track) error
}

// TrackHandler is responsible for performing track type-specific operations
// (eg: saving an MP3, loading a FLAC file, etc.).
type TrackHandler interface {
	Load(ctx context.Context, location string) (*Track, error)
	Save(ctx context.Context, track *Track) (*Track, error)
}
