package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordEncrypt 加密密码
func PasswordEncrypt(password string) string {
	encrypted, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(encrypted)
}

// PasswordCompare 比较密码
func PasswordCompare(password, encrypted string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encrypted), []byte(password))
	return err == nil
}