package models

import "errors"

var (
	ErrInvalidRequest                 = errors.New("invalid request")
	ErrStateMismatch                  = errors.New("state mismatch")
	ErrInvalidAction                  = errors.New("invalid action (only 'signup' and 'login' are allowed)")
	ErrUserAlreadyExists              = errors.New("user already exists")
	ErrUserNotExists                  = errors.New("user not exists")
	ErrTokenNotExists                 = errors.New("no spotify access, try to login again")
	ErrConfirmationTimeout            = errors.New("confirmation timeout")
	ErrCreatePlaylistProcessNotExists = errors.New("no create playlist process running with playlist id: %s")
	ErrCreatePlaylistServiceTimeout   = errors.New("create playlist service timeout")
)
