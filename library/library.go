package library

// Library handles metadata about your media library
type Library struct {
	AudioShelves []AudioShelf
}

// NewLibrary creates a new library
func NewLibrary(m []AudioShelf) (*Library, error) {
	return &Library{
		AudioShelves: m,
	}, nil
}
