package library

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/dhowden/tag"
	log "github.com/sirupsen/logrus"
)

// LocalAudioShelf contains audio media stored in a local filesystem.
type LocalAudioShelf struct {
	directory   string
	files       []string
	filePattern *regexp.Regexp
	tracks      []Track
}

// NewLocalAudioShelf creates a shelf for a specific directory.
func NewLocalAudioShelf(directory string) (*LocalAudioShelf, error) {
	r := regexp.MustCompile(`(.*).(mp3|flac|wav|ogg)$`)

	l := LocalAudioShelf{
		directory:   directory,
		filePattern: r,
	}

	return &l, nil
}

// LoadTracks searches through library for files to add to the database.
// TODO: add unit tests for this
func (l *LocalAudioShelf) LoadTracks() (uint64, error) {
	// look for new files
	i, err := l.pathScan()
	if err != nil {
		return i, err
	}
	log.WithField("count", i).Debug("paths scanned")

	// scan metadata
	i, err = l.loadTracks()
	if err != nil {
		return i, err
	}
	log.WithField("count", i).Debug("files scanned for metadata")
	return i, nil
}

// scan library directory for files
func (l *LocalAudioShelf) pathScan() (uint64, error) {
	var scanCount uint64

	err := filepath.Walk(l.directory,
		func(path string, info os.FileInfo, err error) error {
			log.WithField("path", path).Trace("walking path")
			if err != nil {
				log.WithFields(log.Fields{
					"path":  path,
					"error": err,
				}).Error("could not walk path")
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if !l.ShouldInclude(path) {
				log.WithField("path", path).Debug("discarding path")
				return nil
			}

			log.WithField("path", path).Debug("adding path to library")
			l.files = append(l.files, path)
			scanCount++

			return nil
		})

	if err != nil {
		return scanCount, err
	}

	return scanCount, nil
}

// ShouldInclude checks if we should include the file path in
func (l *LocalAudioShelf) ShouldInclude(path string) bool {
	p := strings.ToLower(path)
	match := l.filePattern.Find([]byte(p))

	if match == nil {
		return false
	}

	return true
}

func (l *LocalAudioShelf) loadTracks() (uint64, error) {
	tracks := []Track{}
	ctx := context.Background()

	// TODO: scan for metadata/id3
	var scanCount uint64
	for _, file := range l.files {
		track, err := l.LoadTrack(ctx, file)
		if err != nil {
			log.WithFields(log.Fields{
				"path":  file,
				"error": err,
			}).Error("could not load track")

			continue
		}
		tracks = append(tracks, *track)
		scanCount++
	}

	l.tracks = tracks
	return scanCount, nil
}

// LoadTrack reads in track metadata
func (l *LocalAudioShelf) LoadTrack(ctx context.Context, path string) (*Track, error) {
	h, err := l.handler(ctx, path)
	if err != nil {
		return nil, err
	}

	return h.Load(ctx, path)
}

// SaveTrack saves track metadata
func (l *LocalAudioShelf) SaveTrack(ctx context.Context, prev, track *Track) (*Track, error) {
	h, err := l.handler(ctx, track.Path)
	if err != nil {
		return nil, err
	}

	return h.Save(ctx, track)
}

// DeleteTrack deletes a track from local audio shelf
func (l *LocalAudioShelf) DeleteTrack(ctx context.Context, track *Track) error {
	if track.Path == "" {
		return errors.New("track has no path")
	}

	err := os.Remove(track.Path)

	if err != nil {
		return err
	}

	return nil
}

// handler returns a filetype-specific track handler responsible for
// loading/saving metadata
func (l *LocalAudioShelf) handler(ctx context.Context, path string) (TrackHandler, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp3":
		return &ID3v2Handler{}, nil
	case ".flac", ".ogg":
		return &TagHandler{}, nil
	case ".wav":
		return &WAVHandler{}, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: [%s]: %s", ext, path)
	}
}

// TagHandler uses the tag package
type TagHandler struct{}

// Load returns metadata for for a track using tag package
func (s *TagHandler) Load(ctx context.Context, path string) (*Track, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file [%s]: [%s]", path, err.Error())
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("could not read metadata [%s]: [%s]", path, err.Error())
	}

	trackNumber, trackTotal := m.Track()
	discNumber, discTotal := m.Disc()

	track := Track{
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
		DiscNumber:  discNumber,
		DiscTotal:   discTotal,
		TrackNumber: trackNumber,
		TrackTotal:  trackTotal,
		Composer:    m.Composer(),
		Year:        m.Year(),
		Genre:       m.Genre(),
		Lyrics:      m.Lyrics(),
		Comment:     m.Comment(),
		FileType:    string(m.FileType()),
		Path:        path,
	}
	return &track, nil
}

// Save track metadata
func (s *TagHandler) Save(ctx context.Context, track *Track) (*Track, error) {
	return nil, nil
}

// ID3v2Handler uses the id3v2 package
type ID3v2Handler struct{}

// Load returns metadata for for a track using id3v2 package
func (s *ID3v2Handler) Load(ctx context.Context, path string) (*Track, error) {
	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if t == nil || err != nil {
		return nil, err
	}
	defer t.Close()

	track := Track{
		Title:  t.Title(),
		Artist: t.Artist(),
		Album:  t.Album(),
		//Year:     t.Year(),
		Genre:       t.Genre(),
		FileType:    "MP3",
		Path:        path,
		RatingEmail: "grump",
	}

	// popm
	f := t.GetLastFrame(t.CommonID("Popularimeter"))
	popm, ok := f.(id3v2.PopularimeterFrame)
	if ok {
		track.Rating = popm.Rating
		track.RatingEmail = popm.Email
	}

	return &track, nil
}

// Save track metadata
func (s *ID3v2Handler) Save(ctx context.Context, track *Track) (*Track, error) {
	log.WithFields(log.Fields{
		"track":  track,
		"artist": track.Artist,
		"album":  track.Album,
		"title":  track.Title,
		"rating": track.Rating,
	}).Debug("saving track")

	// compute diff
	tag, err := id3v2.Open(track.Path, id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		return nil, fmt.Errorf("could not open file [%s]: [%s]", track.Path, err.Error())
	}
	defer tag.Close()

	// Text Tags
	tag.SetTitle(track.Title)
	tag.SetAlbum(track.Album)
	tag.SetArtist(track.Artist)

	// POPM
	frame := tag.GetLastFrame(tag.CommonID("Popularimeter"))
	popm, ok := frame.(id3v2.PopularimeterFrame)
	if ok {
		log.WithFields(log.Fields{
			"prevRating": popm.Rating,
			"prevEmail":  popm.Email,
			"path":       track.Path,
		}).Debug("POPM already set")
	}
	log.WithFields(log.Fields{
		"rating": track.Rating,
		"email":  track.RatingEmail,
		"path":   track.Path,
	}).Debug("setting POPM")

	popmFrame := id3v2.PopularimeterFrame{
		Email:   track.RatingEmail,
		Rating:  track.Rating,
		Counter: big.NewInt(int64(track.PlayCount)),
	}
	tag.AddFrame(tag.CommonID("Popularimeter"), popmFrame)

	return track, tag.Save()
}

// WAVHandler scans wav metadata
type WAVHandler struct{}

// Load scans wav metadata
func (s *WAVHandler) Load(ctx context.Context, path string) (*Track, error) {
	track := Track{
		FileType: "WAV",
		Path:     path,
	}
	return &track, nil
}

// Save track metadata
func (s *WAVHandler) Save(ctx context.Context, track *Track) (*Track, error) {
	return nil, nil
}

// Tracks returns playable audio tracks on the shelf
// TODO scan this
func (l *LocalAudioShelf) Tracks() []Track {
	return l.tracks
}
