package main

import (
	"github.com/celso-alexandre/golang-prisma/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":3333")
}
