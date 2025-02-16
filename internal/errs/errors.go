package errs

import "errors"

var ErrInvalidPassword = errors.New("invalid password")

var ErrUserNotFound = errors.New("user not found")

var ErrMerchNotFound = errors.New("merch not found")

var ErrNotEnoughCoins = errors.New("not enough coins")

var ErrNegativeCoins = errors.New("negative number of coins")

var ErrSendCoinsToYourself = errors.New("you can't send coins to yourself")

var ErrInternalServer = errors.New("internal server error")

var ErrCreateUser = errors.New("could not create user")

var ErrInvalidToken = errors.New("invalid token")
