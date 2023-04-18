package mappings

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mredolatti/tf/codigo/common/dtos/jsend"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/refutil"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/middleware"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/models"

	"github.com/gin-gonic/gin"
)

// Controller bundles endpoints used by the user to interact with sources and files
type Controller struct {
	logger log.Interface
	maps   mapper.Interface
}

// New constructs a new controller
func New(logger log.Interface, maps mapper.Interface) *Controller {
	return &Controller{logger: logger, maps: maps}
}

// Register mounts the endpoints exposed by this controller on a route
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/mappings", c.list)
	router.GET("/mappings/:mappingId", c.get)
	router.POST("/mappings", c.create)
	router.PUT("/mappings/:mappingId", c.update)
	router.DELETE("/mappings/:mappingId", c.remove)
}

func (c *Controller) list(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatusJSON(500, jsend.ResponseErrorInSession)
		return
	}

	var query models.MappingQuery
	if path := ctx.Query("path"); path != "" {
		query.Path = refutil.Ref(path)
	}

	forceUpdate := false
	if force := ctx.Query("forceUpdate"); force == "true" {
		forceUpdate = true
	}

	mappings, err := c.maps.Get(ctx.Request.Context(), session.User(), forceUpdate, &query)
	if err != nil {
        var multiSyncErr *mapper.MultiSyncError
        if errors.As(err, &multiSyncErr) {
            c.logger.Error("when fetching mappings: %s", err.Error())
            handleSyncErrors(ctx, multiSyncErr)
            return
        }
		c.logger.Error("[mappings::list] error fetching: %s", err.Error())
		ctx.AbortWithStatusJSON(500, responseErrGettingMappings)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("mappings", formatMappings(mappings), ""))
}

func (c *Controller) get(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatusJSON(500, jsend.ResponseErrorInSession)
		return
	}

	mappingID := ctx.Param("mappingId")
	if mappingID != "" {
		c.logger.Error("error fetching mapping. no id supplied")
		ctx.AbortWithStatusJSON(400, responseNoID)
		return
	}

	mappings, err := c.maps.Get(ctx.Request.Context(), session.User(), false, &models.MappingQuery{ID: refutil.Ref(mappingID)})
	if err != nil {
        var multiSyncErr *mapper.MultiSyncError
        if errors.As(err, &multiSyncErr) {
            c.logger.Error("when fetching mappings: %s", err.Error())
            handleSyncErrors(ctx, multiSyncErr)
            return
        }
		c.logger.Error("error fetching mapping for user %s: %s", session.User(), err)
		ctx.AbortWithStatusJSON(500, responseErrGettingMappings)
		return
	}

	l := len(mappings)
	switch {
	case l < 1: // no se encontro el mapeo
		ctx.AbortWithStatusJSON(404, jsend.NewCustomFailResponse("", "id", "invalid id provided"))
	case l > 1: // mas de un mapeo con mismo id. error interno
		ctx.AbortWithStatusJSON(500, jsend.NewErrorResponse("invalid result when querying mappings internally"))
	case l == 1: // regio
		ctx.JSON(200, jsend.NewSuccessResponse("mapping", formatMapping(mappings[0]), ""))
	}
}

func (c *Controller) create(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatusJSON(500, jsend.ResponseErrorInSession)
		return
	}

	var dto DTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("error parsing json in request body: %s", err)
		ctx.AbortWithStatusJSON(400, jsend.ResponseFailToReadBody)
		return
	}

	updated, err := c.maps.AddPath(ctx.Request.Context(), session.User(), dto.OrganizationName(), dto.ServerName(), dto.Ref(), dto.Path())
	if err != nil {
		c.logger.Error("error adding new mapping for user %s: %s", session.User(), err)
		c.logger.Debug("recieved mapping that failed: %+v", dto)
		ctx.AbortWithStatusJSON(500, responseErrUpdatingMapping)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("mapping", formatMapping(updated), "")) // regio
}

func (c *Controller) update(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatusJSON(500, jsend.ResponseErrorInSession)
		return
	}

	mappingID := ctx.Param("mappingId")
	if mappingID == "" {
		c.logger.Error("error fetching mapping. no id supplied")
		ctx.AbortWithStatusJSON(400, responseNoID)
		return
	}

	var dto UpdateDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("error parsing json in request body: %s", err)
		ctx.AbortWithStatusJSON(400, jsend.ResponseFailToReadBody)
		return
	}

	updated, err := c.maps.UpdatePathByID(ctx.Request.Context(), session.User(), mappingID, dto.Path)
	if err != nil {
		c.logger.Error("error adding new mapping for user %s: %s", session.User(), err)
		c.logger.Debug("recieved mapping that failed: %+v", dto)
		ctx.AbortWithStatusJSON(500, responseErrUpdatingMapping)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("mapping", formatMapping(updated), "")) // regio
}

func (c *Controller) remove(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatusJSON(500, jsend.ResponseErrorInSession)
		return
	}

	mappingID := ctx.Param("mappingId")
	if mappingID == "" {
		c.logger.Error("error updating mapping. no id supplied")
		ctx.AbortWithStatusJSON(400, responseNoID)
		return
	}

	if err = c.maps.ResetPathByID(ctx.Request.Context(), session.User(), mappingID); err != nil {
		c.logger.Error("error deleting mapping for user %s: %s", session.User(), err)
		ctx.AbortWithStatusJSON(500, responseErrUpdatingMapping)
		return
	}

	ctx.JSON(200, jsend.ResponseEmptySuccess)
}

func handleSyncErrors(ctx *gin.Context, err *mapper.MultiSyncError) {
    data := make(map[string]string)
    err.ForEach(func (org, server string, err error) {
        data[fmt.Sprintf("%s::%s", org, server)] = err.Error()
    })

    resp := jsend.NewErrorResponse("synchronization agains the following org/servers has failed.")
    resp.Data = data
    ctx.AbortWithStatusJSON(500, resp)
}

func formatMappings(mappings []models.Mapping) []DTO {
	toRet := make([]DTO, 0, len(mappings))
	for _, mapping := range mappings {
		toRet = append(toRet, formatMapping(mapping))
	}
	return toRet
}

func formatMapping(mapping models.Mapping) DTO {

	path := mapping.Path()
	if strings.HasPrefix(path, "unassigned") {
		path = ""
	}
	return DTO{
		IDField:               mapping.ID(),
		UserIDField:           mapping.UserID(),
		OrganizationNameField: mapping.OrganizationName(),
		ServerNameField:       mapping.ServerName(),
		SizeBytesField:        mapping.SizeBytes(),
		PathField:             path,
		RefField:              mapping.Ref(),
		UpdatedField:          mapping.Updated().Unix(),
	}
}

var (
	responseErrGettingMappings = jsend.NewErrorResponse("internal error collecting mappings for user")
	responseErrUpdatingMapping = jsend.NewErrorResponse("internal error updating specified mappings")
	responseNoID               = jsend.NewCustomFailResponse("", "id", "invalid id provided")
)
