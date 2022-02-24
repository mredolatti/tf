package filemanager

const (
	// EventFileAvailable is sent when a file becomes available for a user, or when a file that's already
	// been available, is updated. This can be either a new file, new permissions for a user, content update, etc
	EventFileAvailable = iota

	// EventFileNotAvailable is sent when a file is no longer available for a user. This can be either dure to a
	// file being removed or
	EventFileNotAvailable
)

// Change bundles properties associated to a change in a file
type Change struct {
	EventType int
	FileRef   string
	User      string
}

// ChangeListener defines the interface to be implemented by those who want to be notified
// whenever there's been a change in a file
type ChangeListener = func(Change)
