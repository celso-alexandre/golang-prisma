package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/celso-alexandre/golang-prisma/db"
	"github.com/celso-alexandre/golang-prisma/middlewares"
	"github.com/celso-alexandre/golang-prisma/utils"
	"github.com/gin-gonic/gin"
)

func disconnectPrisma(client *db.PrismaClient) {
	if err := client.Prisma.Disconnect(); err != nil {
		panic(err)
	}
}

func connectPrisma(client *db.PrismaClient) {
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
}

func prismaErrorToHttpStatus(err error, resourceName string) (int, string) {
	message := err.Error()
	fmt.Println("PrismaError message: ", message)
	if strings.Contains(message, "Unique constraint failed") {
		return http.StatusConflict, resourceName + " already exists"
	}
	if strings.Contains(message, "ErrNotFound") {
		return http.StatusNotFound, resourceName + " not found"
	}
	return http.StatusInternalServerError, "Internal server error"
}

func getUserFromContext(c *gin.Context) {
	payload, err := middlewares.RetrieveAuthPayload(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func main() {
	server := gin.Default()

	server.POST("/signup", func(c *gin.Context) {
		client := db.NewClient()
		connectPrisma(client)
		defer disconnectPrisma(client)

		ctx := context.Background()

		var u db.UserModel
		err := c.BindJSON(&u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			fmt.Println(err)
			return
		}
		u.Password, err = utils.HashPassword(u.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			fmt.Println(err)
			return
		}
		createdUser, err := client.User.CreateOne(
			db.User.Email.Set(u.Email),
			db.User.Password.Set(u.Password),
		).Exec(ctx)
		if err != nil {
			httpStatusCode, msg := prismaErrorToHttpStatus(err, "User")
			c.JSON(httpStatusCode, gin.H{"error": msg})
			fmt.Println(err)
			return
		}
		c.JSON(http.StatusCreated, createdUser)
	})

	server.POST("/login", func(c *gin.Context) {
		client := db.NewClient()
		connectPrisma(client)
		defer disconnectPrisma(client)

		ctx := context.Background()

		var u db.UserModel
		err := c.BindJSON(&u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			fmt.Println(err)
			return
		}

		user, err := client.User.FindUnique(
			db.User.Email.Equals(u.Email),
		).Exec(ctx)
		if err != nil {
			httpStatusCode, msg := prismaErrorToHttpStatus(err, "Email")
			c.JSON(httpStatusCode, gin.H{"error": msg})
			fmt.Println(err)
			return
		}

		err, httpStatus, msg := utils.ComparePassword(user.Password, u.Password)
		if err != nil {
			fmt.Println(err)
			c.JSON(httpStatus, gin.H{"error": msg})
			return
		}

		token, err := utils.GenerateJwtToken(
			&utils.JwtPayload{UserId: (user.ID), Email: user.Email},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	authenticated := server.Group("/")
	authenticated.Use(middlewares.AuthMiddleware())

	authenticated.GET("/me", getUserFromContext)

	server.Run(":3333")
}
