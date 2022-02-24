package files

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/mredolatti/tf/codigo/fileserver/filemanager"

	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/common/log"

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
		ctx.AbortWithStatus(400)
		return
	}

	metas, err := c.fm.ListFileMetadata(user, nil)
	if err != nil {
		c.logger.Error("files.list: failed to fetch file list for user %s: %s", user, err)
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, gin.H{"objects": toFileMetaDTOs(metas)})
}

func (c *Controller) get(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.get: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.get: no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	meta, err := c.fm.GetFileMetadata(user, id)
	if err != nil {
		c.logger.Error("files.get: unable to fetch file metadata: %s", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
		return
	}

	ctx.JSON(200, gin.H{"object": toFileMetaDTO(meta)})
}

func (c *Controller) create(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.create: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		c.logger.Error("files.create: failed to read request body : %s", err)
		ctx.AbortWithStatus(500)
		return
	}
	defer ctx.Request.Body.Close()

	var dto dtos.FileMetadata
	err = json.Unmarshal(body, &dto)
	if err != nil {
		c.logger.Error("files.create: failed to parse josn in request body : %s", err)
		ctx.AbortWithStatus(400)
		return
	}

	meta, err := c.fm.CreateFileMetadata(user, &dto)
	if err != nil {
		c.logger.Error("files.create: unable to create file metadata: %s", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
		return
	}

	ctx.JSON(200, gin.H{"object": toFileMetaDTO(meta)})
}

func (c *Controller) update(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.update: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.update: no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		c.logger.Error("files.update: failed to read request body : %s", err)
		ctx.AbortWithStatus(500)
		return
	}
	defer ctx.Request.Body.Close()

	var dto dtos.FileMetadata
	err = json.Unmarshal(body, &dto)
	if err != nil {
		c.logger.Error("files.update: failed to parse json in request body : %s", err)
		ctx.AbortWithStatus(400)
		return
	}

	meta, err := c.fm.UpdateFileMetadata(user, id, &dto)
	if err != nil {
		c.logger.Error("files.update: unable to update file metadata: %s", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
		return
	}

	ctx.JSON(200, gin.H{"object": toFileMetaDTO(meta)})

}

func (c *Controller) remove(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.remove: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.remove: no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	if err := c.fm.DeleteFileMetadata(user, id); err != nil {
		c.logger.Error("files.remove: error removing file: %w", err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
	}
}

// Contents mangement endpoints

func (c *Controller) getContents(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.contents.get: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.contents.get: no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	file, err := c.fm.GetFileContents(user, id)
	if err != nil {
		c.logger.Error("files.contents.get: error fetching file contents for %s::%s: : %s", user, id, err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
		return
	}

	ctx.Data(200, "application/octet-stream", file)
}

func (c *Controller) updateContents(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.contents.update: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.contents.get: no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		c.logger.Error("files.contents.update: failed to read body: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if err := c.fm.UpdateFileContents(user, id, body); err != nil {
		c.logger.Error("files.contents.update: error fetching file contents for %s::%s: : %s", user, id, err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
		return
	}
}

func (c *Controller) removeContents(ctx *gin.Context) {
	user := ctx.GetString("user")
	if user == "" {
		c.logger.Error("files.contents.remove: received request with no user")
		ctx.AbortWithStatus(400)
		return
	}

	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("files.contents.remove: no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	if err := c.fm.DeleteFileContents(user, id); err != nil {
		c.logger.Error("files.contents.delete: error deleting file contents for %s::%s: : %s", user, id, err)
		if errors.Is(err, filemanager.ErrUnauthorized) {
			ctx.AbortWithStatus(401)
		} else {
			ctx.AbortWithStatus(500)
		}
		return
	}
}
