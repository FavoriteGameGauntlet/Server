package srvpoints_test

import "errors"

var dbError = errors.New("database connection lost")

func ptrInt(i int) *int { return &i }
