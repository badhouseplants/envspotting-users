package hasher

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var errWrongPassword = errors.New("auth: user with this username is not found or password is incorrect")


func convertPassword(password string) []byte {
	return []byte(password)
}

func hashAndSalt(password []byte) string {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		fmt.Print(err) //FIXME
	}
	return string(hash)
}

// Encrypt is for encrypting password and adding salt
func Encrypt(password string) string {
	return hashAndSalt(convertPassword(password))
}

// ComparePasswords compares password and hash
func ComparePasswords(hashedPassword string, plainPassword string) error {
	byteHash := []byte(hashedPassword)
	password := []byte(plainPassword)
	if err := bcrypt.CompareHashAndPassword(byteHash, password); err != nil {
	return errWrongPassword
	}
	return nil
}