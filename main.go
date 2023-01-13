package main

import (
	"github.com/gin-gonic/gin"
)

var r = gin.Default()

func main() {

	hub := NewHub()
	go hub.Run()
	r.GET("/chat", func(c *gin.Context) {
		serveWs(hub, c.Writer, c.Request)
	})
	r.Run(":8080") // listen and serve on
}
