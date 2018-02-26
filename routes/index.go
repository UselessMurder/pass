package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type indexContent struct {
	Nav *navbarContent
}

func GetIndexHandler(c *gin.Context) {
	ic := &indexContent{getNavBarWithState(4)}
	c.HTML(http.StatusOK, "index", ic)
}
