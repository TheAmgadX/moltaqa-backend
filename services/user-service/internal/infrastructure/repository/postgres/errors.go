package repository

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserInput  = errors.New("invalid user input")
	ErrNothingToUpdate   = errors.New("nothing to update")
	ErrInvalidUserId     = errors.New("invalid user id")
	ErrEmptyUserIdSlice  = errors.New("empty users id slice")
)
