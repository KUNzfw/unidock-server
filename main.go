package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/unidock", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Hello World")
	})
	r.Run(":8080")
}
