package files

import (
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	"github.com/gin-gonic/gin"
)

// Controller bundles endpoints used by the user to interact with sources and files
type Controller struct {
	logger log.Interface
	repo   repository.FileRepository
}

// Register mounts the endpoints exposed by this controller on a route
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/files", c.list)
	router.GET("/organizations/:orgId/files", c.listByOrg)
	router.GET("/organizations/:orgId/files/:fileId", c.get)
}

func (c *Controller) list(ctx *gin.Context) {
	userID, ok := controllers.GetUserOrAbort(ctx)
	if !ok {
		return
	}

	files, err := c.repo.List(userID)
	if err != nil {
		c.logger.Error("error fetching files for user %s: %s", userID, err)
		// TODO: Verificar el error
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, formatFiles(files))
}

func (c *Controller) listByOrg(ctx *gin.Context) {
	userID, ok := controllers.GetUserOrAbort(ctx)
	if !ok {
		return
	}

	orgID := ctx.Param("orgId")
	if orgID == "" {
		c.logger.Error("user '%s' requested organization with empty id", userID)
		ctx.AbortWithStatus(400)
		return
	}

	files, err := c.repo.ListByOrg(userID, orgID)
	if err != nil {
		c.logger.Error("error fetching files for user %s on org %s: %s", userID, orgID, err)
		// TODO: Verificar el error
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, formatFiles(files))
}

func (c *Controller) get(ctx *gin.Context) {
	userID, ok := controllers.GetUserOrAbort(ctx)
	if !ok {
		return
	}

	orgID := ctx.Param("orgId")
	if orgID == "" {
		c.logger.Error("user '%s' requested a file witho empty organizationId", userID)
		ctx.AbortWithStatus(400)
		return
	}

	fileID := ctx.Param("fileId")
	if fileID == "" {
		c.logger.Error("user '%s' requested file with empty id", userID)
		ctx.AbortWithStatus(400)
		return
	}

	file, err := c.repo.Get(userID, orgID, fileID)
	if err != nil {
		c.logger.Error("error fetching organization with id '%s' for user %s: %s", orgID, userID, err)
		// TODO: Verificar el error
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, formatFile(file))
}

func formatFiles(files []models.File) []DTO {
	toRet := make([]DTO, 0, len(files))
	for _, file := range files {
		toRet = append(toRet, formatFile(file))
	}
	return toRet
}

func formatFile(file models.File) DTO {
	return DTO{
		ID:        file.ID(),
		ServerID:  file.ServerID(),
		Ref:       file.Ref(),
		Size:      file.Size(),
		PatientID: file.PatientID(),
		Updated:   file.Updated().Unix(),
	}
}
