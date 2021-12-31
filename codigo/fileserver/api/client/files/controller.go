package files

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mredolatti/tf/codigo/fileserver/authz"
	"github.com/mredolatti/tf/codigo/fileserver/storage"

	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
)

// Controller implements file-interaction endpoints
type Controller struct {
	logger        log.Interface
	authorization authz.Authorization
	fileMetas     storage.FilesMetadata
	files         storage.Files
}

// New constructs a new controller
func New(logger log.Interface,
	authorization authz.Authorization,
	files storage.Files,
	metas storage.FilesMetadata,
) *Controller {
	return &Controller{
		logger:        logger,
		authorization: authorization,
		files:         files,
		fileMetas:     metas,
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

	objsWithAuth := c.authorization.AllForSubject(user)
	fileIDList := make([]string, 0, len(objsWithAuth))
	for id := range objsWithAuth {
		fileIDList = append(fileIDList, id)
	}

	metas, err := c.fileMetas.GetMany(fileIDList)
	if err != nil {
		c.logger.Error("files.list: error reading files: %w", err)
		ctx.AbortWithStatus(500)
		return
	}

	result := make([]dtos.FileMetadata, 0, len(metas))
	for _, meta := range metas {
		result = append(result, toFileMetaDTO(meta))
	}

	ctx.JSON(200, gin.H{"objects": result})
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

	allowed, err := c.authorization.Can(user, authz.Read, id)
	if err != nil {
		c.logger.Error("files.get: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.get: user %s is not allowed to read file %s", user, id)
		ctx.AbortWithStatus(403)
		return
	}

	meta, err := c.fileMetas.Get(id)
	if err != nil {
		c.logger.Error("files.get: error reading file: %w", err)
		ctx.AbortWithStatus(500)
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

	allowed, err := c.authorization.Can(user, authz.Create, authz.AnyObject)
	if err != nil {
		c.logger.Error("files.create: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.create: user %s is not allowed to create a new fille", user)
		ctx.AbortWithStatus(403)
		return
	}

	meta, err := c.fileMetas.Create(dto.PName, dto.PNotes, dto.PPatientID, dto.PType)
	if err != nil {
		c.logger.Error("files.create: error creating file: %w", err)
		ctx.AbortWithStatus(500)
		return
	}

	c.authorization.Grant(user, authz.Read, meta.ID())
	c.authorization.Grant(user, authz.Write, meta.ID())
	c.authorization.Grant(user, authz.Admin, meta.ID())
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

	allowed, err := c.authorization.Can(user, authz.Write, id)
	if err != nil {
		c.logger.Error("files.update: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.update: user %s is not allowed to update a file", user)
		ctx.AbortWithStatus(403)
		return
	}

	meta, err := c.fileMetas.Update(id, &dto)
	if err != nil {
		c.logger.Error("files.update: error updating file: %w", err)
		ctx.AbortWithStatus(500)
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

	allowed, err := c.authorization.Can(user, authz.Write, id)
	if err != nil {
		c.logger.Error("files.remove: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.remove: user %s is not allowed to removing file %s", user, id)
		ctx.AbortWithStatus(403)
		return
	}

	err = c.fileMetas.Remove(id)
	if err != nil {
		c.logger.Error("files.remove: error removing file: %w", err)
		ctx.AbortWithStatus(500)
		return
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

	allowed, err := c.authorization.Can(user, authz.Read, id)
	if err != nil {
		c.logger.Error("files.contents.get: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.contents.get: user %s is not allowed to read file %s", user, id)
		ctx.AbortWithStatus(403)
		return
	}

	file, err := c.files.Read(id)
	if err != nil {
		c.logger.Error("files.contents.get: failed when reading item %s from storage: %s", id, err)
		ctx.AbortWithStatus(500)
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

	allowed, err := c.authorization.Can(user, authz.Create, authz.AnyObject)
	if err != nil {
		c.logger.Error("files.contents.update: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.contents.update: user %s is not allowed to create files", user)
		ctx.AbortWithStatus(403)
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

	allowed, err := c.authorization.Can(user, authz.Read, id)
	if err != nil {
		c.logger.Error("files.contents.remove: failed to get permission: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !allowed {
		c.logger.Error("files.contents.remove: user %s is not allowed to remove file contents for %s", user, id)
		ctx.AbortWithStatus(403)
		return
	}

	c.files.Del(id)

}
