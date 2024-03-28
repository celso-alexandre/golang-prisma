package utils

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func ComparePassword(hashedPassword, password string) (error, int, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return err, http.StatusUnauthorized, "Invalid password"
		}
		return err, http.StatusInternalServerError, "Internal server error"
	}
	return nil, http.StatusOK, ""
}
