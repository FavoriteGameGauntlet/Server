package validator_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/validator"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr[T any](value T) *T {
	return &value
}

func TestValidator_ValidateEffectUseCount(test *testing.T) {
	tests := []struct {
		name      string
		useCount  *int
		wantError bool
	}{
		{name: "Nil_NoLimit", useCount: nil},
		{name: "One", useCount: ptr(1)},
		{name: "IntegerMaximum", useCount: ptr(math.MaxInt32)},
		{name: "Zero_Rejected", useCount: ptr(0), wantError: true},
		{name: "Negative_Rejected", useCount: ptr(-3), wantError: true},
		{name: "AboveInteger_Rejected", useCount: ptr(math.MaxInt32 + 1), wantError: true},
	}

	for _, tt := range tests {
		test.Run(tt.name, func(test *testing.T) {
			err := validator.ValidateEffectUseCount(tt.useCount)

			if !tt.wantError {
				require.NoError(test, err)
				return
			}

			var unprocessable *common.UnprocessableError
			require.ErrorAs(test, err, &unprocessable)
			require.Equal(test, "INCORRECT_EFFECT_USE_COUNT", unprocessable.Code)
		})
	}
}

func TestValidator_ValidateEffectDurationInSeconds(test *testing.T) {
	tests := []struct {
		name      string
		seconds   *int
		wantError bool
	}{
		{name: "Nil_NeverRunsOut", seconds: nil},
		{name: "OneSecond", seconds: ptr(1)},
		{name: "LongestFittingDuration", seconds: ptr(int(math.MaxInt64 / 1_000_000_000))},
		{name: "Zero_Rejected", seconds: ptr(0), wantError: true},
		{name: "Negative_Rejected", seconds: ptr(-60), wantError: true},
		{name: "OverflowingDuration_Rejected", seconds: ptr(int(math.MaxInt64/1_000_000_000) + 1), wantError: true},
	}

	for _, tt := range tests {
		test.Run(tt.name, func(test *testing.T) {
			err := validator.ValidateEffectDurationInSeconds(tt.seconds)

			if !tt.wantError {
				require.NoError(test, err)
				return
			}

			var unprocessable *common.UnprocessableError
			require.ErrorAs(test, err, &unprocessable)
			require.Equal(test, "INCORRECT_EFFECT_DURATION", unprocessable.Code)
		})
	}
}

func TestValidator_ValidateEffectModifierAmount(test *testing.T) {
	tests := []struct {
		name      string
		amount    int
		wantError bool
	}{
		{name: "Positive", amount: 3},
		{name: "Negative", amount: -3},
		{name: "IntegerMaximum", amount: math.MaxInt32},
		{name: "IntegerMinimum", amount: math.MinInt32},
		{name: "Zero_Rejected", amount: 0, wantError: true},
		{name: "AboveInteger_Rejected", amount: math.MaxInt32 + 1, wantError: true},
		{name: "BelowInteger_Rejected", amount: math.MinInt32 - 1, wantError: true},
	}

	for _, tt := range tests {
		test.Run(tt.name, func(test *testing.T) {
			err := validator.ValidateEffectModifierAmount(tt.amount)

			if !tt.wantError {
				require.NoError(test, err)
				return
			}

			var unprocessable *common.UnprocessableError
			require.ErrorAs(test, err, &unprocessable)
			require.Equal(test, "INCORRECT_EFFECT_MODIFIER_AMOUNT", unprocessable.Code)
		})
	}
}
