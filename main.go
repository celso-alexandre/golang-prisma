package main

import (
	"context"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/celso-alexandre/golang-prisma/db"
)

func main() {
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	ctx := context.Background()

	createdUser, err := client.User.CreateOne(
		db.User.Email.Set(gofakeit.Email()),
		db.User.Password.Set(gofakeit.Password(true, true, true, true, false, 14)),
	).Exec(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(createdUser)
}
