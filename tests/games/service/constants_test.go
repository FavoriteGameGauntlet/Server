package srvgames_test

import (
	"errors"
	"time"
)

var dbError = errors.New("database connection lost")

var gameStartDate = time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
