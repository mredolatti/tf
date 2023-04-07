package organizations

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/dtos/jsend"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/middleware"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

type Controller struct {
	logger    log.Interface
	registrar registrar.Interface
}

func New(registrar registrar.Interface, logger log.Interface) *Controller {
	return &Controller{
		logger:    logger,
		registrar: registrar,
	}
}

func (c *Controller) Register(router gin.IRouter) {
	router.GET("/organizations", c.listOrganizations)
	router.GET("/organizations/:name", c.getOrganization)
	router.GET("/organizations/:name/servers", c.listServersForOrg)
	router.GET("/organizations/:name/servers/:serverName", c.listServersForOrg)
	router.GET("/organizations/:name/servers/:serverName/link", c.initiateLinkProcess)
	router.GET("/servers", c.listServers)
	router.GET("/servers/:serverId", c.getServer)
}

func (c *Controller) listOrganizations(ctx *gin.Context) {
	orgs, err := c.registrar.ListOrganizations(ctx.Request.Context())
	if err != nil {
		c.logger.Error("error fetching organizations: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("organizations", toOrgsView(orgs), ""))
}

func (c *Controller) getOrganization(ctx *gin.Context) {
	org, err := c.registrar.GetOrganization(ctx.Request.Context(), ctx.Param("name"))
	switch err {
	case nil:
	case repository.ErrNotFound:
		ctx.AbortWithStatus(404)
		return
	default:
		c.logger.Error("error fetching organization: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("organization", toOrgView(org), ""))
}

func (c *Controller) listServersForOrg(ctx *gin.Context) {
	id := ctx.Param("name")
	fss, err := c.registrar.ListServers(ctx.Request.Context(), models.FileServersQuery{OrganizationName: &id})
	if err != nil {
		c.logger.Error("error fetching servers for organization %s: %s", id, err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("servers", toFileServersView(fss), ""))
}

func (c *Controller) listServers(ctx *gin.Context) {
	fss, err := c.registrar.ListServers(ctx.Request.Context(), models.FileServersQuery{})
	if err != nil {
		c.logger.Error("error fetching servers: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, jsend.NewSuccessResponse("servers", toFileServersView(fss), ""))
}

func (c *Controller) getServer(ctx *gin.Context) {
	// TODO(mredolatti: validate org id?
	server, err := c.registrar.GetServer(ctx.Request.Context(), ctx.Param("name"), ctx.Param("serverName"))
	switch err {
	case nil:
	case repository.ErrNotFound:
		ctx.AbortWithStatus(404)
		return
	default:
		c.logger.Error("error fetching organization: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}
    ctx.JSON(200, jsend.NewSuccessResponse("server", toFileServerView(server), ""))
}

func (c *Controller) initiateLinkProcess(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session information: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	orgName := ctx.Param("name")
	serverName := ctx.Param("serverName")
	force := ctx.Query("force") == "true"

	url, err := c.registrar.InitiateLinkProcess(ctx.Request.Context(), session.User(), orgName, serverName, force)
	if err != nil {
		if errors.Is(err, registrar.ErrAccountExists) {
			ctx.JSON(400, "account already exists")
			c.logger.Error("requested initial link with an already existing account (%s/%s/%s)", session.User(), orgName, serverName)
			return
		}
		ctx.JSON(500, "unable to initiate oauth2 flow")
		c.logger.Error("error initiating oauth2 flow: %s", err)
		return
	}

	ctx.Redirect(301, url)
}

func toOrgView(org models.Organization) OrganizationViewDTO {
	return OrganizationViewDTO{
		ID:   org.ID(),
		Name: org.Name(),
	}
}

func toOrgsView(orgs []models.Organization) []OrganizationViewDTO {
	res := make([]OrganizationViewDTO, len(orgs))
	for i := range orgs {
		res[i] = toOrgView(orgs[i])
	}
	return res
}

func toFileServerView(fs models.FileServer) FileServerViewDTO {
	return FileServerViewDTO{
		ID:                fs.ID(),
		OrganizationName:  fs.OrganizationName(),
		Name:              fs.Name(),
		AuthenticationURL: fs.AuthURL(),
		TokenURL:          fs.TokenURL(),
		FileFetchURL:      fs.FetchURL(),
		ControlEndpoint:   fs.ControlEndpoint(),
	}
}

func toFileServersView(servers []models.FileServer) []FileServerViewDTO {
	res := make([]FileServerViewDTO, len(servers))
	for i := range servers {
		res[i] = toFileServerView(servers[i])
	}
	return res
}
