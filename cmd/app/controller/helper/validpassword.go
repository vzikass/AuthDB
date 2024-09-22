package helper

import (
	"regexp"
	"unicode"
)

func IsNumeric(s string) bool {
	// numbers from 0 to 9 matches the previous token between one and unlimited times
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

// Checking the password for valid letters and digits
func IsValidPassword(password string) bool {
	var hasDigit, hasLetter bool
	if len(password) < 4 {
		return false
	}

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
		if hasDigit && hasLetter {
			return true
		}
	}

	return hasDigit && hasLetter
}
