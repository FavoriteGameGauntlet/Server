package srvwheeleffects_test

import "errors"

var dbError = errors.New("database connection lost")

func ptr[T any](v T) *T { return &v }
