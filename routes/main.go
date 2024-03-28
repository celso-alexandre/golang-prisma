package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/celso-alexandre/golang-prisma/db"
	"github.com/celso-alexandre/golang-prisma/middlewares"
	"github.com/celso-alexandre/golang-prisma/utils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	authenticated := server.Group("/")
	authenticated.Use(middlewares.AuthMiddleware())

	server.POST("/signup", signup)
	server.POST("/login", login)

	authenticated.GET("/me", getUserFromContext)
}

func getUserFromContext(c *gin.Context) {
	payload, err := middlewares.RetrieveAuthPayload(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func signup(c *gin.Context) {
	client := db.NewClient()
	db.ConnectPrisma(client)
	defer db.DisconnectPrisma(client)

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
		httpStatusCode, msg := db.PrismaErrorToHttpStatus(err, "User")
		c.JSON(httpStatusCode, gin.H{"error": msg})
		fmt.Println(err)
		return
	}
	c.JSON(http.StatusCreated, createdUser)
}

func login(c *gin.Context) {
	client := db.NewClient()
	db.ConnectPrisma(client)
	defer db.DisconnectPrisma(client)

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
		httpStatusCode, msg := db.PrismaErrorToHttpStatus(err, "Email")
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
		&utils.JwtPayload{UserId: user.ID, Email: user.Email},
	)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
