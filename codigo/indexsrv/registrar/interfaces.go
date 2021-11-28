package registrar

import "github.com/mredolatti/tf/codigo/indexsrv/models"

// AdminInterface of a registrar
// This interface specifies methods that only admin have access such CRUD operations on an org,
// or (de)associating file-servers.
type AdminInterface interface {
	RegisterOrganization(name string) (models.Organization, error)
	RegisterServer(server models.FileServer) (models.FileServer, error)
}

// UserInterface of a registrar
// This interface specifies methods that can be called by regular users
type UserInterface interface {
	ListOrganizations() ([]models.Organization, error)
	GetOrganization(id string) (models.Organization, error)
	Link(userID string, orgID string) error
}
