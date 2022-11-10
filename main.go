package main

import (
	"faktur/controller"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	engine.POST("/compute", controller.ComputeData)

	engine.SetTrustedProxies(nil)
	engine.Run(":8080")
}
