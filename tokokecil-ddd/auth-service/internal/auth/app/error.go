package app

import "errors"

var (
	ErrValidation   = errors.New("validasi gagal")
	ErrEmailExist   = errors.New("email sudah terdaftar")
	ErrUnauthorized = errors.New("email/password salah")
)
