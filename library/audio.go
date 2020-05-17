package library

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dhowden/tag"
	log "github.com/sirupsen/logrus"
)

// AudioShelf is an abstract collection of audio. A shelf has one source type
// (local, internet, spotify account, etc.). For example, a LocalAudioShelf
// contains files stored in a local filesystem.
type AudioShelf interface {
	Tracks() []Track

	// Scan searches for new files to add to the library
	Scan() (uint64, error)
}

// Track represents audio media from any source
type Track struct {
	Artist      string
	Album       string
	Title       string
	Path        string
	Rating      int
	AlbumArtist string
	DiscNumber  int
	DiscTotal   int
	TrackNumber int
	TrackTotal  int
	Year        int
	Composer    string
	Genre       string
	Lyrics      string
	Comment     string

	FileType string
	MimeType string
}

// LocalAudioShelf contains audio media stored in a local filesystem.
type LocalAudioShelf struct {
	directory   string
	files       []string
	filePattern *regexp.Regexp
	tracks      []Track
}

func NewLocalAudioShelf(directory string) (*LocalAudioShelf, error) {
	r := regexp.MustCompile(`(.*).(mp3|flac|wav|ogg)$`)

	l := LocalAudioShelf{
		directory:   directory,
		filePattern: r,
	}

	return &l, nil
}

// Scan searches through library for files to add to the database.
// TODO: add unit tests for this
func (l *LocalAudioShelf) Scan() (uint64, error) {
	// look for new files
	i, err := l.pathScan()
	if err != nil {
		return i, err
	}
	log.WithField("count", i).Debug("paths scanned")

	// scan metadata
	i, err = l.metadataScan()
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
			log.WithField("path", path).Debug("walking path")
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

func (l *LocalAudioShelf) metadataScan() (uint64, error) {
	tracks := []Track{}
	ctx := context.Background()

	// TODO: scan for metadata/id3
	var scanCount uint64
	for _, file := range l.files {
		ext := strings.ToLower(filepath.Ext(file))
		var scanner TrackScanner

		switch ext {
		case ".mp3", ".flac", ".ogg":
			scanner = &TagScanner{}
		case ".wav":
			scanner = &WAVScanner{}
		default:
			log.WithFields(log.Fields{
				"path": file,
				"ext":  ext,
			}).Error("unsupported file extension")
			continue
		}

		track, err := scanner.Scan(ctx, file)
		if err != nil {
			log.WithFields(log.Fields{
				"path":  file,
				"error": err,
			}).Error("could not scan metadata")

			continue
		}
		tracks = append(tracks, *track)
		scanCount++
	}

	l.tracks = tracks
	return scanCount, nil
}

// TrackScanner scans something and returns a track
type TrackScanner interface {
	Scan(context.Context, string) (*Track, error)
}

// TagScanner uses the tag package
type TagScanner struct{}

// Scan returns metadata for for a track using tag package
func (s *TagScanner) Scan(ctx context.Context, path string) (*Track, error) {
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

type WAVScanner struct{}

func (s *WAVScanner) Scan(ctx context.Context, path string) (*Track, error) {
	track := Track{
		Title:    path,
		FileType: "WAV",
		Path:     path,
	}
	return &track, nil
}

// TODO scan this
func (l *LocalAudioShelf) Tracks() []Track {
	return l.tracks
}
