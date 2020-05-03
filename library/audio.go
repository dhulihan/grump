package library

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/dhowden/tag"
	"github.com/sirupsen/logrus"
)

// AudioShelf is an abstract collection of audio. A shelf has one source type
// (local, internet, spotify account, etc.). For example, a LocalAudioShelf
// contains files stored in a local filesystem.
type AudioShelf interface {
	Directory() string
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
	r := regexp.MustCompile("(.*).[mp3|flac]$")

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
	logrus.WithField("count", i).Debug("paths scanned")

	// scan metadata
	i, err = l.metadataScan()
	if err != nil {
		return i, err
	}
	logrus.WithField("count", i).Debug("files scanned for metadata")
	return i, nil
}

// scan library directory for files
func (l *LocalAudioShelf) pathScan() (uint64, error) {
	var scanCount uint64

	err := filepath.Walk(l.directory,
		func(path string, info os.FileInfo, err error) error {
			logrus.WithField("path", path).Debug("walking path")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"path":  path,
					"error": err,
				}).Error("could not walk path")
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if !l.shouldInclude(path) {
				logrus.WithField("path", path).Debug("discarding path")
				return nil
			}

			logrus.WithField("path", path).Debug("adding path to library")
			l.files = append(l.files, path)
			scanCount++

			return nil
		})

	if err != nil {
		return scanCount, err
	}

	return scanCount, nil
}

func (l *LocalAudioShelf) shouldInclude(path string) bool {
	match := l.filePattern.Find([]byte(path))

	if match == nil {
		return false
	}

	return true
}

func (l *LocalAudioShelf) Directory() string {
	return l.directory
}

func (l *LocalAudioShelf) metadataScan() (uint64, error) {
	tracks := []Track{}

	// TODO: scan for metadata/id3
	var scanCount uint64
	for _, file := range l.files {
		f, err := os.Open(file)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path":  file,
				"error": err,
			}).Error("could not open file")
			continue
		}

		m, err := tag.ReadFrom(f)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path":  file,
				"error": err,
			}).Error("could not read metadata")
			continue
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
			Path:        file,
		}
		tracks = append(tracks, track)
		scanCount++
	}

	l.tracks = tracks
	return scanCount, nil
}

// TODO scan this
func (l *LocalAudioShelf) Tracks() []Track {
	return l.tracks
}
