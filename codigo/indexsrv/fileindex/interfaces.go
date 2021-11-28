package fileindex

import "github.com/mredolatti/tf/codigo/indexsrv/models"

// AdminInterface of a file-index
// This interface specifies the set of methods that only admins can invoke in order to
// publish a set of file updates (additions/updates & removals), and restrict permissions
// to a certain set of users
type AdminInterface interface {
	Publish(orgID string, serverID string, added []models.File, removed []models.File, checkpoint int64) error
	Restrict(users []models.User, files []models.File, read bool, write bool) error
}

// UserInterface of a file-index
// This interface allows users (health-care workers) to retrieve a set of files for which he has access
// pre-filtered with a query.
type UserInterface interface {
	Get(userID string, query *models.FileQuery) ([]models.File, error)
}
