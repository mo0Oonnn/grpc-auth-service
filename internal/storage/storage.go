package storage

import "fmt"

var (
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrAppNotFound       = fmt.Errorf("app not found")
	ErrNotFound          = fmt.Errorf("not found")
)
