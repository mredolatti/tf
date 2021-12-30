package dtos

// FileMetadata contains information associated to a file
type FileMetadata struct {
	PID        string `json:"id"`
	PName      string `json:"name"`
	PNotes     string `json:"notes"`
	PPatientID string `json:"patientId"`
	PType      string `json:"type"`
	PContentID string `json:"contentId"`
}

// ID returns the id of the file meta
func (f *FileMetadata) ID() string {
	return f.PID
}

// Name returns the name of the file
func (f *FileMetadata) Name() string {
	return f.PName
}

// Notes returns associated notes
func (f *FileMetadata) Notes() string {
	return f.PNotes
}

// PatientID returnst the id of the patient this file is associated with
func (f *FileMetadata) PatientID() string {
	return f.PPatientID
}

// Type returns the type of file
func (f *FileMetadata) Type() string {
	return f.PType
}

// ContentID returns the id of the associated content entry
func (f *FileMetadata) ContentID() string {
	return f.PContentID
}
