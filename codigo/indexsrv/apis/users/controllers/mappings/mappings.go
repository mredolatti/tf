package mappings

import (
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
	// 	router.POST("/mappings", c.create)
	// 	router.PUT("/mappings/:mappingId", c.update)
	// 	router.DELETE("/mappings/:mappingId", c.list)
}

func (c *Controller) list(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	var query models.MappingQuery
	if path := ctx.Query("path"); path != "" {
		query.Path = refutil.Ref(path)
	}

	forceUpdate := false
	mappings, err := c.maps.Get(ctx.Request.Context(), session.User(), forceUpdate, &query)
	if err != nil {
		c.logger.Error("[mappings::list] error fetching: %s", err.Error())
		ctx.AbortWithStatusJSON(500, "error fetching mappings")
		return
	}

	resp, err := jsend.NewSuccessResponse("mapping", formatMappings(mappings), "")
	if err != nil {
		c.logger.Error("[mappings::list] error building response: %s", err.Error())
		ctx.AbortWithStatusJSON(500, "error building response")
		return
	}
	ctx.JSON(200, resp)
}

func (c *Controller) get(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session data: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	mappingID := ctx.Param("mappingId")
	if mappingID != "" {
		c.logger.Error("error fetching mapping. no id supplied")
		ctx.AbortWithStatus(400)
		return
	}

	mappings, err := c.maps.Get(ctx.Request.Context(), session.User(), false, &models.MappingQuery{ID: refutil.Ref(mappingID)})
	if err != nil {
		c.logger.Error("error fetching mappings for user %s: %s", session.User(), err)
		ctx.AbortWithStatus(500)
		return
	}

	l := len(mappings)
	switch {
	case l < 1:
		ctx.AbortWithStatus(404) // no se encontro el mapeo
	case l > 1:
		ctx.AbortWithStatus(500) // mas de un mapeo con mismo id. error interno
	case l == 1:
		ctx.JSON(200, formatMapping(mappings[0])) // regio
	}
}

// func (c *Controller) create(ctx *gin.Context) {
// 	userID, ok := controllers.GetUserOrAbort(ctx)
// 	if !ok {
// 		return
// 	}
//
// 	body, err := ioutil.ReadAll(ctx.Request.Body)
// 	if err != nil {
// 		c.logger.Error("error reading request body: %s", err)
// 		ctx.AbortWithStatus(400)
// 		return
// 	}
//
// 	var dto DTO
// 	err = json.Unmarshal(body, &dto)
// 	if err != nil {
// 		c.logger.Error("error parsing json in request body: %s", err)
// 		ctx.AbortWithStatus(400)
// 		return
// 	}
//
// 	added, err := c.repo.Add(ctx.Request.Context(), userID, &dto)
// 	if err != nil {
// 		c.logger.Error("error adding new mapping for user %s: %s", userID, err)
// 		c.logger.Debug("recieved mapping that failed: %+v", dto)
// 		ctx.AbortWithStatus(500)
// 		return
// 	}
//
// 	ctx.JSON(200, added)
// }
//
// func (c *Controller) update(ctx *gin.Context) {
// 	userID, ok := controllers.GetUserOrAbort(ctx)
// 	if !ok {
// 		return
// 	}
//
// 	mappingID := ctx.Param("mappingId")
// 	if mappingID != "" {
// 		c.logger.Error("error updating mapping. no id supplied")
// 		ctx.AbortWithStatus(500)
// 		return
// 	}
//
// 	body, err := ioutil.ReadAll(ctx.Request.Body)
// 	if err != nil {
// 		c.logger.Error("error reading request body: %s", err)
// 		ctx.AbortWithStatus(400)
// 		return
// 	}
//
// 	var dto DTO
// 	err = json.Unmarshal(body, &dto)
// 	if err != nil {
// 		c.logger.Error("error parsing json in request body: %s", err)
// 		ctx.AbortWithStatus(400)
// 		return
// 	}
//
// 	added, err := c.repo.Update(ctx.Request.Context(), userID, mappingID, &dto)
// 	if err != nil {
// 		c.logger.Error("error updting mapping for user %s: %s", userID, err)
// 		c.logger.Debug("recieved mapping that failed: %+v", dto)
// 		ctx.AbortWithStatus(500)
// 		return
// 	}
//
// 	ctx.JSON(200, added)
// }
//
// func (c *Controller) remove(ctx *gin.Context) {
// 	userID, ok := controllers.GetUserOrAbort(ctx)
// 	if !ok {
// 		return
// 	}
//
// 	mappingID := ctx.Param("mappingId")
// 	if mappingID != "" {
// 		c.logger.Error("error updating mapping. no id supplied")
// 		ctx.AbortWithStatus(500)
// 		return
// 	}
//
// 	err := c.repo.Remove(ctx.Request.Context(), userID, mappingID)
// 	if err != nil {
// 		c.logger.Error("error deleting mapping for user %s: %s", userID, err)
// 		ctx.AbortWithStatus(500)
// 		return
// 	}
//
// 	ctx.JSON(200, "")
// }
//
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
		UserIDField:           mapping.UserID(),
		OrganizationNameField: mapping.OrganizationName(),
		ServerNameField:       mapping.ServerName(),
		SizeBytesField:        mapping.SizeBytes(),
		PathField:             path,
		RefField:              mapping.Ref(),
		UpdatedField:          mapping.Updated().Unix(),
	}
}
