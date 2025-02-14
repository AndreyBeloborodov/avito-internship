package errs

import "errors"

var ErrInvalidPassword = errors.New("invalid password")

var ErrUserNotFound = errors.New("user not found")

var ErrNotEnoughCoins = errors.New("not enough coins")

var ErrNegativeCoins = errors.New("negative number of coins")
