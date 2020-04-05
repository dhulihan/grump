package library

type MockAudioLibrary struct {
}

func NewMockAudioLibrary() (*MockAudioLibrary, error) {
	l := MockAudioLibrary{}

	return &l, nil
}

func (l *MockAudioLibrary) Directory() string {
	return "no directory available"
}

func (l *MockAudioLibrary) Tracks() []Track {
	track := Track{
		Title:  "Mock Track",
		Artist: "Mock Artist",
		Album:  "Mock Album",
		Path:   "sample-audio/pizza-pie.mp3",
	}
	return []Track{track}
}
