package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are defined and have proper messages
	assert.Equal(t, "not found", ErrNotFound.Error())
	assert.Equal(t, "invalid input", ErrInvalidInput.Error())
	assert.Equal(t, "internal error", ErrInternal.Error())
}

func TestWError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		// Create a WError without wrapping an error
		werr := &WError{
			msg: "something went wrong",
		}
		assert.Equal(t, "something went wrong", werr.Error())
	})

	t.Run("with wrapped error", func(t *testing.T) {
		// Create a WError wrapping another error
		werr := &WError{
			msg: "failed to process",
			err: ErrNotFound,
		}
		assert.Equal(t, "failed to process: not found", werr.Error())
	})

	t.Run("with wrapped WError", func(t *testing.T) {
		// Nested wrapping
		inner := &WError{
			msg: "database error",
		}
		outer := &WError{
			msg: "service failed",
			err: inner,
		}
		assert.Equal(t, "service failed: database error", outer.Error())
	})

	t.Run("with wrapped standard error", func(t *testing.T) {
		stdErr := errors.New("original error")
		werr := &WError{
			msg: "wrapped",
			err: stdErr,
		}
		assert.Equal(t, "wrapped: original error", werr.Error())
	})
}

func TestWError_Unwrap(t *testing.T) {
	t.Run("returns underlying error", func(t *testing.T) {
		werr := &WError{
			msg: "wrapper",
			err: ErrInvalidInput,
		}
		assert.Equal(t, ErrInvalidInput, werr.Unwrap())
	})

	t.Run("returns nil when no wrapped error", func(t *testing.T) {
		werr := &WError{
			msg: "standalone error",
		}
		assert.Nil(t, werr.Unwrap())
	})

	t.Run("returns nested wrapped error", func(t *testing.T) {
		inner := &WError{
			msg: "inner",
			err: ErrNotFound,
		}
		outer := &WError{
			msg: "outer",
			err: inner,
		}
		// Outer Unwrap returns the inner WError
		assert.Equal(t, inner, outer.Unwrap())
		// Inner Unwrap returns ErrNotFound
		assert.Equal(t, ErrNotFound, inner.Unwrap())
	})

	t.Run("supports errors.Is", func(t *testing.T) {
		werr := &WError{
			msg: "wrapper",
			err: ErrNotFound,
		}
		assert.True(t, errors.Is(werr, ErrNotFound))
	})
}

func TestWError_WithFields(t *testing.T) {
	t.Run("attaches fields to empty WError", func(t *testing.T) {
		werr := &WError{
			msg: "test error",
		}
		result := werr.WithFields(map[string]any{
			"key1": "value1",
			"key2": 123,
		})
		assert.Equal(t, "value1", result.fields["key1"])
		assert.Equal(t, 123, result.fields["key2"])
	})

	t.Run("merges fields with existing fields", func(t *testing.T) {
		werr := &WError{
			msg: "test error",
			fields: map[string]any{
				"existing": "field",
			},
		}
		result := werr.WithFields(map[string]any{
			"new": "field",
		})
		assert.Equal(t, "field", result.fields["existing"])
		assert.Equal(t, "field", result.fields["new"])
	})

	t.Run("overwrites existing field", func(t *testing.T) {
		werr := &WError{
			msg: "test error",
			fields: map[string]any{
				"key": "old value",
			},
		}
		result := werr.WithFields(map[string]any{
			"key": "new value",
		})
		assert.Equal(t, "new value", result.fields["key"])
	})

	t.Run("returns same WError instance", func(t *testing.T) {
		werr := &WError{
			msg: "test error",
		}
		result := werr.WithFields(map[string]any{"key": "value"})
		assert.Same(t, werr, result)
	})

	t.Run("nil fields map initializes correctly", func(t *testing.T) {
		werr := &WError{
			msg: "test error",
		}
		assert.Nil(t, werr.fields)
		result := werr.WithFields(map[string]any{"key": "value"})
		assert.NotNil(t, result.fields)
	})
}

func TestWrap(t *testing.T) {
	t.Run("returns nil when err is nil", func(t *testing.T) {
		result := Wrap(nil, "some message")
		assert.Nil(t, result)
	})

	t.Run("creates WError with message", func(t *testing.T) {
		result := Wrap(ErrNotFound, "resource operation")
		assert.NotNil(t, result)
		werr, ok := result.(*WError)
		assert.True(t, ok)
		assert.Equal(t, "resource operation", werr.msg)
		assert.Equal(t, ErrNotFound, werr.err)
	})

	t.Run("creates WError with standard error", func(t *testing.T) {
		stdErr := errors.New("original error")
		result := Wrap(stdErr, "wrapped message")
		werr := result.(*WError)
		assert.Equal(t, "wrapped message", werr.msg)
		assert.Equal(t, stdErr, werr.err)
	})

	t.Run("empty message is allowed", func(t *testing.T) {
		result := Wrap(ErrInvalidInput, "")
		werr := result.(*WError)
		assert.Equal(t, "", werr.msg)
	})

	t.Run("preserves error chain", func(t *testing.T) {
		err1 := errors.New("level 1")
		err2 := Wrap(err1, "level 2")
		err3 := Wrap(err2, "level 3")

		assert.Equal(t, err2, err3.(*WError).err)
		assert.Equal(t, err1, err2.(*WError).err)
	})
}

func TestWrapWithFields(t *testing.T) {
	t.Run("returns nil when err is nil", func(t *testing.T) {
		result := WrapWithFields(nil, "message", map[string]any{"key": "value"})
		assert.Nil(t, result)
	})

	t.Run("creates WError with message and fields", func(t *testing.T) {
		fields := map[string]any{"request_id": "123", "user_id": 456}
		result := WrapWithFields(ErrNotFound, "resource not found", fields)

		werr, ok := result.(*WError)
		assert.True(t, ok)
		assert.Equal(t, "resource not found", werr.msg)
		assert.Equal(t, ErrNotFound, werr.err)
		assert.Equal(t, "123", werr.fields["request_id"])
		assert.Equal(t, 456, werr.fields["user_id"])
	})

	t.Run("nil fields is allowed", func(t *testing.T) {
		result := WrapWithFields(ErrInternal, "error", nil)
		werr := result.(*WError)
		assert.NotNil(t, werr)
		assert.Nil(t, werr.fields)
	})
}

func TestIsNotFound(t *testing.T) {
	t.Run("returns true for ErrNotFound", func(t *testing.T) {
		assert.True(t, IsNotFound(ErrNotFound))
	})

	t.Run("returns true for wrapped ErrNotFound", func(t *testing.T) {
		werr := &WError{
			msg: "wrapped",
			err: ErrNotFound,
		}
		assert.True(t, IsNotFound(werr))
	})

	t.Run("returns true for nested wrapped ErrNotFound", func(t *testing.T) {
		inner := &WError{
			msg: "inner",
			err: ErrNotFound,
		}
		outer := &WError{
			msg: "outer",
			err: inner,
		}
		assert.True(t, IsNotFound(outer))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		assert.False(t, IsNotFound(ErrInvalidInput))
		assert.False(t, IsNotFound(ErrInternal))
		assert.False(t, IsNotFound(errors.New("some error")))
	})

	t.Run("returns false for nil", func(t *testing.T) {
		assert.False(t, IsNotFound(nil))
	})

	t.Run("returns true via NewNotFound", func(t *testing.T) {
		assert.True(t, IsNotFound(NewNotFound("item not found")))
	})
}

func TestIsInvalidInput(t *testing.T) {
	t.Run("returns true for ErrInvalidInput", func(t *testing.T) {
		assert.True(t, IsInvalidInput(ErrInvalidInput))
	})

	t.Run("returns true for wrapped ErrInvalidInput", func(t *testing.T) {
		werr := &WError{
			msg: "wrapped",
			err: ErrInvalidInput,
		}
		assert.True(t, IsInvalidInput(werr))
	})

	t.Run("returns true for nested wrapped ErrInvalidInput", func(t *testing.T) {
		inner := &WError{
			msg: "inner",
			err: ErrInvalidInput,
		}
		outer := &WError{
			msg: "outer",
			err: inner,
		}
		assert.True(t, IsInvalidInput(outer))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		assert.False(t, IsInvalidInput(ErrNotFound))
		assert.False(t, IsInvalidInput(ErrInternal))
	})

	t.Run("returns true via NewInvalidInput", func(t *testing.T) {
		assert.True(t, IsInvalidInput(NewInvalidInput("invalid data")))
	})
}

func TestIsInternal(t *testing.T) {
	t.Run("returns true for ErrInternal", func(t *testing.T) {
		assert.True(t, IsInternal(ErrInternal))
	})

	t.Run("returns true for wrapped ErrInternal", func(t *testing.T) {
		werr := &WError{
			msg: "wrapped",
			err: ErrInternal,
		}
		assert.True(t, IsInternal(werr))
	})

	t.Run("returns true for nested wrapped ErrInternal", func(t *testing.T) {
		inner := &WError{
			msg: "inner",
			err: ErrInternal,
		}
		outer := &WError{
			msg: "outer",
			err: inner,
		}
		assert.True(t, IsInternal(outer))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		assert.False(t, IsInternal(ErrNotFound))
		assert.False(t, IsInternal(ErrInvalidInput))
	})

	t.Run("returns true via NewInternal", func(t *testing.T) {
		assert.True(t, IsInternal(NewInternal("service failure")))
	})
}

func TestIsErrorMatch(t *testing.T) {
	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, isErrorMatch(nil, ErrNotFound))
	})

	t.Run("detects direct sentinel match", func(t *testing.T) {
		assert.True(t, isErrorMatch(ErrNotFound, ErrNotFound))
	})

	t.Run("detects sentinel match in wrapped WError", func(t *testing.T) {
		werr := &WError{
			msg: "wrapper",
			err: ErrNotFound,
		}
		assert.True(t, isErrorMatch(werr, ErrNotFound))
	})

	t.Run("detects sentinel match in nested WError", func(t *testing.T) {
		inner := &WError{
			msg: "inner",
			err: ErrNotFound,
		}
		outer := &WError{
			msg: "outer",
			err: inner,
		}
		assert.True(t, isErrorMatch(outer, ErrNotFound))
	})

	t.Run("returns false for non-matching sentinel", func(t *testing.T) {
		werr := &WError{
			msg: "wrapper",
			err: ErrNotFound,
		}
		assert.False(t, isErrorMatch(werr, ErrInternal))
	})

	t.Run("detects message content match for non-WError", func(t *testing.T) {
		// Only standard errors (not WError) match via message content
		stdErr := errors.New("internal error occurred")
		assert.True(t, isErrorMatch(stdErr, ErrInternal))
	})

	t.Run("standard error with message match", func(t *testing.T) {
		stdErr := errors.New("not found: user")
		assert.True(t, isErrorMatch(stdErr, ErrNotFound))
	})

	t.Run("deeply nested chain", func(t *testing.T) {
		err := ErrNotFound
		for i := 0; i < 5; i++ {
			err = Wrap(err, "level")
		}
		assert.True(t, isErrorMatch(err, ErrNotFound))
	})
}

func TestNewInvalidInput(t *testing.T) {
	t.Run("creates error with message (includes wrapped sentinel)", func(t *testing.T) {
		// NewInvalidInput wraps ErrInvalidInput, so Error() includes both messages
		err := NewInvalidInput("email is required")
		assert.Equal(t, "email is required: invalid input", err.Error())
	})

	t.Run("wraps ErrInvalidInput", func(t *testing.T) {
		err := NewInvalidInput("invalid field")
		assert.True(t, errors.Is(err, ErrInvalidInput))
	})

	t.Run("IsInvalidInput returns true", func(t *testing.T) {
		err := NewInvalidInput("some validation error")
		assert.True(t, IsInvalidInput(err))
	})

	t.Run("Unwrap returns ErrInvalidInput", func(t *testing.T) {
		err := NewInvalidInput("test")
		werr := err.(*WError)
		assert.Equal(t, ErrInvalidInput, werr.Unwrap())
	})
}

func TestNewNotFound(t *testing.T) {
	t.Run("creates error with message (includes wrapped sentinel)", func(t *testing.T) {
		// NewNotFound wraps ErrNotFound, so Error() includes both messages
		err := NewNotFound("user with id 123")
		assert.Equal(t, "user with id 123: not found", err.Error())
	})

	t.Run("wraps ErrNotFound", func(t *testing.T) {
		err := NewNotFound("resource")
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	t.Run("IsNotFound returns true", func(t *testing.T) {
		err := NewNotFound("item not found")
		assert.True(t, IsNotFound(err))
	})

	t.Run("Unwrap returns ErrNotFound", func(t *testing.T) {
		err := NewNotFound("test")
		werr := err.(*WError)
		assert.Equal(t, ErrNotFound, werr.Unwrap())
	})
}

func TestNewInternal(t *testing.T) {
	t.Run("creates error with message (includes wrapped sentinel)", func(t *testing.T) {
		// NewInternal wraps ErrInternal, so Error() includes both messages
		err := NewInternal("database connection failed")
		assert.Equal(t, "database connection failed: internal error", err.Error())
	})

	t.Run("wraps ErrInternal", func(t *testing.T) {
		err := NewInternal("something broke")
		assert.True(t, errors.Is(err, ErrInternal))
	})

	t.Run("IsInternal returns true", func(t *testing.T) {
		err := NewInternal("system error")
		assert.True(t, IsInternal(err))
	})

	t.Run("Unwrap returns ErrInternal", func(t *testing.T) {
		err := NewInternal("test")
		werr := err.(*WError)
		assert.Equal(t, ErrInternal, werr.Unwrap())
	})
}

func TestErrorInterface(t *testing.T) {
	t.Run("WError implements error interface", func(t *testing.T) {
		var err error = &WError{msg: "test"}
		assert.NotNil(t, err)
		assert.Equal(t, "test", err.Error())
	})

	t.Run("Wrap result implements error interface", func(t *testing.T) {
		var err error = Wrap(ErrNotFound, "test")
		assert.NotNil(t, err)
	})

	t.Run("New* results implement error interface", func(t *testing.T) {
		var err error = NewNotFound("test")
		assert.NotNil(t, err)
	})
}

func TestWrapChaining(t *testing.T) {
	t.Run("full error chain", func(t *testing.T) {
		// Simulate: service -> repository -> database
		dbErr := NewInternal("connection refused")
		repoErr := Wrap(dbErr, "failed to fetch user")
		svcErr := Wrap(repoErr, "user service unavailable")

		// Check the chain
		assert.Contains(t, svcErr.Error(), "user service unavailable")
		assert.Contains(t, svcErr.Error(), "failed to fetch user")
		assert.Contains(t, svcErr.Error(), "connection refused")

		// All sentinel checks should pass through
		assert.True(t, IsInternal(svcErr))
		assert.True(t, IsInternal(repoErr))
		assert.True(t, IsInternal(dbErr))
	})
}
