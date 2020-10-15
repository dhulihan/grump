package library

// Track represents audio media from any source
type Track struct {
	Album       string
	AlbumArtist string
	Artist      string
	Comment     string
	Composer    string
	DiscNumber  int
	DiscTotal   int
	FileType    string
	Genre       string

	// Length is length of track in millis
	Length      int
	Lyrics      string
	MimeType    string
	Path        string
	PlayCount   uint64
	Rating      uint8
	RatingEmail string
	Title       string
	TrackNumber int
	TrackTotal  int
	Year        int
}

func (t Track) String() string {
	return t.Path
}
