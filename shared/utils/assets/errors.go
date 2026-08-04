package assets

import "errors"

var (
	ErrInvalidProfileImagePath = errors.New("invalid profile image path")
	ErrInvalidProfileImageSize = errors.New("invalid profile image size")
	ErrInvalidProfileImageType = errors.New("invalid profile image type")
)
