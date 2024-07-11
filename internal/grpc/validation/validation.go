package validation

import (
	"strings"

	"github.com/asaskevich/govalidator"
)

const (
	emptyID = 0
)

func IsEmail(email string) bool {
	return govalidator.IsEmail(email)
}

func IsValidAppID(appID int32) bool {
	return appID > emptyID
}

func IsValidUserID(userID int64) bool {
	return userID > emptyID
}

func IsValidPassword(password string) bool {
	return CheckLength(password, 5) && IsRequired(password)
}

func CheckLength(value string, min int) bool {
	return len([]rune(value)) >= min
}

func IsRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}
