package service

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNoLiquidity  = errors.New("no liquidity")
)
