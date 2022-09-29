package domain

import "errors"

var (
	ErrUserNotFound                = errors.New("user doesn't exists")
	ErrUserAlreadyExists           = errors.New("user with such email already exists")
	ErrPasswordIncorrect           = errors.New("password incorrect")
	ErrAuthorizationInvalidToken   = errors.New("token invalid")
	ErrSecretNoSecretForUser       = errors.New("no secret for the user")
	ErrSecretFieldsAreNotValid     = errors.New("either secret or created_at field is not empty/invalid")
	ErrSecretNotValid              = errors.New("error while scanning secret")
	ErrSecretAccessesCountExceeded = errors.New("attempts threshold to get secret exceeded")
	ErrSecretAccessesCountInValid  = errors.New("error when getting count accesses")
	ErrAccountFieldsInValid        = errors.New("some account fields are invalid")
	ErrAccountExpired              = errors.New("refresh token for account has expired")
	ErrAutorizationSigningMethod   = errors.New("unexpected signing method")
	ErrSecretHasExpired            = errors.New("unexpected signing method")
)
