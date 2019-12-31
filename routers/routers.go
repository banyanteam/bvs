package routers

import (

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/bvs/utils"
)


var Router *gin.Engine


type APIHandler struct {
	RestartChan chan bool
}

var API = &APIHandler{
	RestartChan: make(chan bool),
}


func init() {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = utils.GetLogWriter()
}


func Init() error {
	
	Router = gin.New()
	Router.Use(gin.Recovery())


	// Ping test
	Router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})


	api := Router.Group("/av-domain-service/v1")
	api.POST("/stream", API.StreamService)
	api.POST("/devicegw", API.DeviceGateway)

	return nil
}
