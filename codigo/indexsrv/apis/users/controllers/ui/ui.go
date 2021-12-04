package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/frontend"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Controller serves endpoints that render ui pages
type Controller struct {
	logger log.Interface
	mapper mapper.Interface
}

// New constructs a new UI controller
func New(logger log.Interface, mapper mapper.Interface) *Controller {
	return &Controller{
		logger: logger,
		mapper: mapper,
	}
}

// Register mounts the endpoints onto the supplied router
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/", c.main)
	router.GET("/main", c.main)
	router.GET("/main/mappings", c.mappings)
}

func (c *Controller) main(ctx *gin.Context) {

	session := sessions.Default(ctx)
	if id := session.Get("id"); id == nil {
		ctx.Header("Content-Type", "text/html")
		ctx.String(200, string(frontend.Login()))
		return
	}

	// the user's logged in, render the page
	ctx.Header("Content-Type", "text/html")
	ctx.String(200, string(frontend.Index()))
}

func (c *Controller) mappings(ctx *gin.Context) {
	// session := sessions.Default(ctx)
	// if id := session.Get("id"); id == nil {
	// 	ctx.AbortWithStatus(401)
	// 	return
	// }

	// id := session.Get("id").(string)
	id := "107156877088323945674"
	mappings, err := c.mapper.Get(ctx.Request.Context(), id, nil) // TODO: Buildear query aca
	if err != nil {
		c.logger.Error("error fetching mappings for user %s: %s", id, err)
		ctx.AbortWithStatus(500)
	}

	fmt.Println("AA", mappings)
	ctx.JSON(200, formatMappings(mappings))
}

type nodes []node

type node struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Type     string `json:"type"`
	Children nodes  `json:"children"`
}

func (n *nodes) lookupChild(text string) *node {

	/*
		for idx := range *n {
			if (*n)[idx].Text == text {
				return &(*n)[idx]
			}
		}
	*/

	i, j := 0, len(*n)
	for i < j {
		currIndex := int(uint(i+j) >> 1) // avoid overflow when computing h

		curr := (*n)[currIndex].Text
		if curr == text {
			return &(*n)[currIndex]
		} else if curr < text {
			i = currIndex + 1
		} else {
			j = currIndex
		}
	}
	return nil
}

func (n *nodes) getOrAdd(id string, text string, t string) *node {
	if found := n.lookupChild(text); found != nil {
		return found
	}
	(*n) = append(*n, node{ID: id, Text: text, Type: t})
	return &(*n)[len(*n)-1]
}

func formatMappings(mappings []models.Mapping) []node {
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Path() < mappings[j].Path()
	})

	var n nodes
	i := 0
	for _, mapping := range mappings {
		i++
		pathComponents := strings.Split(mapping.Path(), "/")
		if len(pathComponents) == 0 { // odd... skip
			continue
		}

		curr := n.getOrAdd(strconv.Itoa(i), pathComponents[0], "folder")
		for _, pathComponent := range pathComponents[1 : len(pathComponents)-1] {
			i++
			curr = curr.Children.getOrAdd(strconv.Itoa(i), pathComponent, "folder")
		}
		i++
		curr.Children.getOrAdd(strconv.Itoa(i), pathComponents[len(pathComponents)-1], "file")

	}

	return n
}
