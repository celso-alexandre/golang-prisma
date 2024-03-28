package db

import (
	"fmt"
	"net/http"
	"strings"
)

func DisconnectPrisma(client *PrismaClient) {
	if err := client.Prisma.Disconnect(); err != nil {
		panic(err)
	}
}

func ConnectPrisma(client *PrismaClient) {
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
}

func PrismaErrorToHttpStatus(err error, resourceName string) (int, string) {
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
