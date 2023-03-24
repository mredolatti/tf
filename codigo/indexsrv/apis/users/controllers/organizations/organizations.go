package organizations

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/refutil"
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
		logger: logger,
		registrar: registrar,
	}
}

func (c *Controller) Register(router gin.IRouter) {
	router.GET("/organizations", c.listOrganizations)
	router.GET("/organizations/:id", c.getOrganization)
	router.GET("/organizations/:id/servers", c.listServersForOrg)
	router.GET("/servers", c.listServers)
	router.GET("/servers/:serverId", c.getServer)
	router.GET("/servers/:serverId/link", c.initiateLinkProcess) // hacer redirect a => GET en file server
}

func (c *Controller) listOrganizations(ctx *gin.Context) {
	orgs, err := c.registrar.ListOrganizations(ctx.Request.Context())
	if err != nil {
		c.logger.Error("error fetching organizations: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, toOrgsView(orgs))
}

func (c *Controller) getOrganization(ctx *gin.Context) {
	org, err := c.registrar.GetOrganization(ctx.Request.Context(), ctx.Param("id"))
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
	ctx.JSON(200, toOrgView(org))
}

func (c *Controller) listServersForOrg(ctx *gin.Context) {
	fss, err := c.registrar.ListServers(ctx.Request.Context(), models.FileServersQuery{OrgID: refutil.Ref(ctx.Param("id"))})
	if err != nil {
		c.logger.Error("error fetching organizations: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, toFileServersView(fss, "")) // TODO(mredolatti): pass proper name or remove

}

func (c *Controller) listServers(ctx *gin.Context) {
	fss, err := c.registrar.ListServers(ctx.Request.Context(), models.FileServersQuery{IDs: ctx.QueryArray("id")})
	if err != nil {
		c.logger.Error("error fetching organizations: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, toFileServersView(fss, "")) // TODO(mredolatti): pass proper name or remove

}

func (c *Controller) getServer(ctx *gin.Context) {
	// TODO(mredolatti: validate org id?
	server, err := c.registrar.GetServer(ctx.Request.Context(), ctx.Param("serverId"))
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
	ctx.JSON(200, toFileServerView(server, ""))
}

func (c *Controller) initiateLinkProcess(ctx *gin.Context) {

	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Error("error getting session information: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	serverID := ctx.Param("serverId")
	force := ctx.Query("force") == "true"

	url, err := c.registrar.InitiateLinkProcess(ctx.Request.Context(), session.User(), serverID, force)
	if err != nil {
		if errors.Is(err, registrar.ErrAccountExists) {
			ctx.JSON(400, "account already exists")
			c.logger.Error("requested initial link with an already existing account (%s/%s)", session.User(), serverID)
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

func toFileServerView(fs models.FileServer, orgName string) FileServerViewDTO {
	return FileServerViewDTO{
		ID:                fs.ID(),
		OrganizationName:  orgName,
		Name:              fs.Name(),
		AuthenticationURL: fs.AuthURL(),
		TokenURL:          fs.TokenURL(),
		FileFetchURL:      fs.FetchURL(),
		ControlEndpoint:   fs.ControlEndpoint(),
	}
}

func toFileServersView(servers []models.FileServer, orgName string) []FileServerViewDTO {
	res := make([]FileServerViewDTO, len(servers))
	for i := range servers {
		res[i] = toFileServerView(servers[i], orgName)
	}
	return res
}
