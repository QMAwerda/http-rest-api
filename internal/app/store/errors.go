package store

import "errors"

// Тут будут лежать ошибки, которые мы будем возвращать (стандартизация, чтоб наши ошибки и ошибки postgres звучали одинаково)

var (
	ErrRecordNotFound = errors.New("record not found")
)
