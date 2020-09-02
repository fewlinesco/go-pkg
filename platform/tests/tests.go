package tests

import (
	"fmt"
	"testing"
	"time"
)

// AssertString ensures both strings are equal or fails the test
func AssertString(t *testing.T, expected string, received string, errorMessage string) {
	if expected != received {
		message := fmt.Sprintf("\t%s. \nExpected: %v, \nhave:  %v", errorMessage, expected, received)
		t.Fatalf(message, expected, received)
	}
}

// AssertTimestamp ensures both time are equal or fails the test
func AssertTimestamp(t *testing.T, expected time.Time, received time.Time, errorMessage string) {
	if !expected.Equal(received) {
		message := fmt.Sprintf("\t%s. \nExpected: %v, \nhave:  %v", errorMessage, expected, received)
		t.Fatalf(message, expected, received)
	}
}

// AssertBoolean ensures both boolean are equal or fails the test
func AssertBoolean(t *testing.T, expected bool, received bool, errorMessage string) {
	if expected != received {
		message := fmt.Sprintf("\t%s. \nExpected: %v, \nhave:  %v", errorMessage, expected, received)
		t.Fatalf(message, expected, received)
	}
}
