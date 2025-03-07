package payment_test

import (
	"os"
	"testing"
)

func init() {
	// Set the APP_ENV to test for all tests in this package
	os.Setenv("APP_ENV", "test")
}

// TestMain is used to set up any test-wide configuration
func TestMain(m *testing.M) {
	// Run the tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}
