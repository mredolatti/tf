package admin

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/dtos/jsend"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
)

// Controller bundles endpoints used by the user to interact with sources and files
type Controller struct {
	logger log.Interface
	registrar registrar.Interface
}

func New(registrar registrar.Interface, logger log.Interface) *Controller {
	return &Controller{
		logger: logger,
		registrar: registrar,
	}
}

// Register mounts the endpoints exposed by this controller on a route
func (c *Controller) Register(router gin.IRouter) {
	router.POST("/organizations", c.create)
	router.GET("/organizations", c.list)
	router.GET("/organizations/:orgId", c.get)
}

func (c *Controller) create(ctx *gin.Context) {
	var dto DTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("error reading body: ", err)
		ctx.AbortWithStatusJSON(400, jsend.NewReadBodyFailResponse(err))
		return
	}

	if err := c.registrar.AddNewOrganization(ctx.Request.Context(), dto.Name); err != nil {
		c.logger.Error("error creating new organization: ", err)
		ctx.AbortWithStatusJSON(500, jsend.NewErrorResponse("error creating organization"))
		return
	}

	ctx.JSON(200, jsend.ResponseEmptySuccess)
}

func (c *Controller) list(ctx *gin.Context) {
	userID, ok := controllers.GetUserOrAbort(ctx)
	if !ok {
		return
	}

	orgs, err := c.registrar.ListOrganizations(ctx.Request.Context())
	if err != nil {
		c.logger.Error("error fetching organizations for user %s: %s", userID, err)
		// TODO: Verificar el error
		ctx.AbortWithStatusJSON(500, jsend.NewErrorResponse("error reading organizations from db"))
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("organizations", formatOrgs(orgs), ""))
}

func (c *Controller) get(ctx *gin.Context) {
	userID, ok := controllers.GetUserOrAbort(ctx)
	if !ok {
		return
	}

	orgID := ctx.Param("orgId")
	if orgID == "" {
		c.logger.Error("user '%s' requested organization with empty id", userID)
		ctx.AbortWithStatusJSON(400, jsend.NewCustomFailResponse("invalid input", "id", "cannot be empty/null"))
		return
	}

	org, err := c.registrar.GetOrganization(context.Background(), orgID)
	if err != nil {
		c.logger.Error("error fetching organization with id '%s' for user %s: %s", orgID, userID, err)
		// TODO: Verificar el error
		ctx.AbortWithStatusJSON(500, jsend.NewErrorResponse("error fetching organization from db"))
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("organization", formatOrg(org), ""))
}

func formatOrgs(orgs []models.Organization) []DTO {
	toRet := make([]DTO, 0, len(orgs))
	for _, org := range orgs {
		toRet = append(toRet, formatOrg(org))
	}
	return toRet
}

func formatOrg(org models.Organization) DTO {
	return DTO{
		ID:   org.ID(),
		Name: org.Name(),
	}
}
