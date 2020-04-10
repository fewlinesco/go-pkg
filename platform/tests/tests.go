package tests

import (
	"fmt"
	"testing"
	"time"
)

func AssertString(testProvider *testing.T, expected string, received string, errorMessage string) {
	if expected != received {
		message := fmt.Sprintf("\t%s. \nExpected: %v, \nhave:  %v", errorMessage, expected, received)
		testProvider.Fatalf(message, expected, received)
	}
}

func AssertTimestamp(testProvider *testing.T, expected time.Time, received time.Time, errorMessage string) {
	if !expected.Equal(received) {
		message := fmt.Sprintf("\t%s. \nExpected: %v, \nhave:  %v", errorMessage, expected, received)
		testProvider.Fatalf(message, expected, received)
	}
}

func AssertBoolean(testProvider *testing.T, expected bool, received bool, errorMessage string) {
	if expected != received {
		message := fmt.Sprintf("\t%s. \nExpected: %v, \nhave:  %v", errorMessage, expected, received)
		testProvider.Fatalf(message, expected, received)
	}
}
