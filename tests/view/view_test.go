package view_test

import (
	"testing"

	"craftops/internal/view"
)

func TestPrintBanner(t *testing.T) {
	// This is a simple test to ensure the function runs without panicking.
	// A more complex test would involve capturing stdout, which is overkill here.
	view.PrintBanner("Test Banner")
}
