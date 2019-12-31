package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gin-gonic/gin"
)

func auth(c *gin.Context) {
	data, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Printf("ctx.Request.body: %v\n", string(data))
	c.String(http.StatusOK, "{\"code\": 1, \"data\": \"ok\"}")
}

func main() {
	router := gin.Default()
	router.POST("/streamcloud-control-service/srs/appgw/auth", auth)	
	router.Run(":30004")


	// userGroup := router.Group("/user")
	// {
	// 	userGroup.POST("/login", userLogin)
	// }

	// realGroup := router.Group("/real")
	// {
	// 	realGroup.POST("/get", hello)
	// 	realGroup.POST("/play", hello)
	// 	realGroup.POST("/stop", hello)
	// }

	// vodGroup := router.Group("/vod")
	// {
	// 	vodGroup.POST("/query", hello)
	// 	vodGroup.POST("/play", hello)
	// 	vodGroup.POST("/stop", hello)
	// }

	// // Listen and Server in https://127.0.0.1:8080
	// router.RunTLS(":8080", "../testdata/server.pem", "../testdata/server.key")

}

