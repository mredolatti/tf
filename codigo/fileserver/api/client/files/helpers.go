package files

import (
	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/fileserver/models"
)

func toFileMetaDTO(meta models.FileMetadata) dtos.FileMetadata {
	return dtos.FileMetadata{
		PID:          meta.ID(),
		PName:        meta.Name(),
		PNotes:       meta.Notes(),
		PPatientID:   meta.PatientID(),
		PType:        meta.Type(),
		PContentID:   meta.ContentID(),
		PLastUpdated: meta.LastUpdated(),
	}
}

func toFileMetaDTOs(metas []models.FileMetadata) []dtos.FileMetadata {
	result := make([]dtos.FileMetadata, 0, len(metas))
	for _, meta := range metas {
		result = append(result, toFileMetaDTO(meta))
	}
	return result
}
