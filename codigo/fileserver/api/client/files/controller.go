package files

import (
	"errors"
	"io/ioutil"

	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/common/dtos/jsend"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"

	"github.com/gin-gonic/gin"
)

// Controller implements file-interaction endpoints
type Controller struct {
	logger log.Interface
	fm     filemanager.Interface
}

// New constructs a new controller
func New(logger log.Interface, manager filemanager.Interface) *Controller {
	return &Controller{
		logger: logger,
		fm:     manager,
	}
}

// Register mounts the login endpoints onto the supplied router
func (c *Controller) Register(router gin.IRouter) {

	// File record metadata
	router.GET("/files", c.list)
	router.GET("/files/:id", c.get)
	router.POST("/files", c.create)
	router.PUT("/files/:id", c.update)
	router.DELETE("/files/:id", c.remove)

	// File contents
	router.GET("/files/:id/contents", c.getContents)
	router.PUT("/files/:id/contents", c.updateContents)
	router.DELETE("/files/:id/contents", c.removeContents)
}

func (c *Controller) list(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.list: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	metas, err := c.fm.ListFileMetadata(user, nil)
	if err != nil {
		c.logger.Error("files.list: failed to fetch file list for user %s: %s", user, err)
		ctx.AbortWithStatusJSON(500, responseErrorFetchingMetadata)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("files", toFileMetaDTOs(metas), ""))
}

func (c *Controller) get(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.get: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.get: no id supplied")
		ctx.AbortWithStatusJSON(400, responseFailNoID)
		return
	}

	meta, err := c.fm.GetFileMetadata(user, id)
	if err != nil {
		c.logger.Error("files.get: unable to fetch file metadata: %s", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorFetchingMetadata)
		}
		return
	}

	ctx.JSON(200, gin.H{"object": toFileMetaDTO(meta)})
}

func (c *Controller) create(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.create: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	var dto dtos.FileMetadata
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("files.create: failed to parse josn in request body : %s", err)
		ctx.AbortWithStatusJSON(400, jsend.NewReadBodyFailResponse(err))
		return
	}

	meta, err := c.fm.CreateFileMetadata(user, &dto)
	if err != nil {
		c.logger.Error("files.create: unable to create file metadata: %s", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorWritingMetadata)
		}
		return
	}

	ctx.JSON(200, gin.H{"object": toFileMetaDTO(meta)})
}

func (c *Controller) update(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.update: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.update: no id supplied")
		ctx.AbortWithStatusJSON(400, responseFailNoID)
		return
	}

	var dto dtos.FileMetadata
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("files.update: failed to parse json in request body : %s", err)
		ctx.AbortWithStatusJSON(400, jsend.NewReadBodyFailResponse(err))
		return
	}

	meta, err := c.fm.UpdateFileMetadata(user, id, &dto)
	if err != nil {
		c.logger.Error("files.update: unable to update file metadata: %s", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorWritingMetadata)
		}
		return
	}

	ctx.JSON(200, gin.H{"object": toFileMetaDTO(meta)})

}

func (c *Controller) remove(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.remove: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.remove: no id supplied")
		ctx.AbortWithStatusJSON(400, responseFailNoID)
		return
	}

	if err := c.fm.DeleteFileMetadata(user, id); err != nil {
		c.logger.Error("files.remove: error removing file: %w", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorWritingMetadata)
		}
	}

    ctx.JSON(200, jsend.ResponseEmptySuccess)

}

// Contents mangement endpoints

func (c *Controller) getContents(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.contents.get: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.contents.get: no id supplied")
		ctx.AbortWithStatusJSON(400, responseFailNoID)
		return
	}

	file, err := c.fm.GetFileContents(user, id)
	if err != nil {
		c.logger.Error("files.contents.get: error fetching file contents for %s::%s: : %s", user, id, err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorFetchingContents)
		}
		return
	}

	ctx.Data(200, "application/octet-stream", file)
}

func (c *Controller) updateContents(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.contents.update: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.contents.get: no id supplied")
		ctx.AbortWithStatusJSON(400, responseFailNoID)
		return
	}

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		c.logger.Error("files.contents.update: failed to read body: %s", err)
		ctx.AbortWithStatusJSON(500, responseErrorWritingContents)
		return
	}

	if err := c.fm.UpdateFileContents(user, id, body); err != nil {
		c.logger.Error("files.contents.update: error updating file contents for [%s::%s] : %s", user, id, err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorWritingContents)
		}
		return
	}

    ctx.JSON(200, jsend.ResponseEmptySuccess)

}

func (c *Controller) removeContents(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.contents.remove: received request with no user")
		ctx.AbortWithStatusJSON(500, responseNoUser)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.contents.remove: no id supplied")
		ctx.AbortWithStatusJSON(400, responseFailNoID)
		return
	}

	if err := c.fm.DeleteFileContents(user, id); err != nil {
		c.logger.Error("files.contents.delete: error deleting file contents for %s::%s: : %s", user, id, err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatusJSON(401, responseUnauthorized)
		} else {
			ctx.AbortWithStatusJSON(500, responseErrorWritingContents)
		}
		return
	}

    ctx.JSON(200, jsend.ResponseEmptySuccess)
}

var (
	responseNoUser                = jsend.NewErrorResponse("internal error processing client authentication")
	responseErrorFetchingMetadata = jsend.NewErrorResponse("internal error fetching files information")
	responseErrorWritingMetadata  = jsend.NewErrorResponse("internal error writing file information")
	responseErrorFetchingContents = jsend.NewErrorResponse("internal error fetching file contents")
	responseErrorWritingContents  = jsend.NewErrorResponse("internal error writing file contents")
	responseFailNoID              = jsend.NewCustomFailResponse("", "id", "parameter is mandatory and missing")
	responseUnauthorized          = jsend.NewCustomFailResponse("", "reason", "insufficient permissions")
)
