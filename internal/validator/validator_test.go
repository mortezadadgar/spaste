package validator

import (
	"fmt"
	"testing"
)

func TestIsEqual(t *testing.T) {
	const data = "data"
	var expectedErr = fmt.Errorf("%s: %s: %s", validationErr, data, isBlankErr)

	t.Run("returns no error", func(t *testing.T) {
		validator := New()

		validator.IsBlank(data, "test")
		validator.IsEqual(data, 0, 1)

		got := validator.Valid()

		assertNoError(t, got)
	})

	t.Run("returns expected error", func(t *testing.T) {
		validator := New()

		validator.IsBlank(data, "")
		validator.IsEqual(data, 0, 1)

		want := expectedErr
		got := validator.Valid()

		assertError(t, got, want)
	})
}

func assertNoError(t *testing.T, got error) {
	t.Helper()
	if got != nil {
		t.Errorf("didn't expect a error but got one")
	}
}

func assertError(t *testing.T, got error, want error) {
	t.Helper()
	if got == nil {
		t.Errorf("didn't get error but wanted one")
	}

	if got.Error() != want.Error() {
		t.Errorf("got %v, want %v", got, want)
	}
}

var err error

func BenchmarkValidator(b *testing.B) {
	errChannel := make(chan error)
	validator := New()
	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		go func() {
			validator.IsBlank("data", "")
			validator.IsEqual("data", 0, 1)

			errChannel <- validator.Valid()
		}()
	}

	err = <-errChannel
}
