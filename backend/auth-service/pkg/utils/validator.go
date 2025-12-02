package utils

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) bool {
	return len(password) >= 6
}

func ValidateVerificationCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
