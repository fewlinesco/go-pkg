package retry

import (
	"net/http"
	"time"
)

// Config describes the configuration options for the middleware
type Config struct {
	MaxRetry int
	Delay    time.Duration
	ExceptOn []int
}

type retryRoundTripper struct {
	roundTripper http.RoundTripper
	retryConfig  Config
}

func (retryRoundTripper retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	doRequest := func(request *http.Request) (*http.Response, error) {
		return retryRoundTripper.roundTripper.RoundTrip(request)
	}

	maxReTries := retryRoundTripper.retryConfig.MaxRetry

	response, err := doRequest(req)
	if err != nil || !isExceptStatus(response.StatusCode, retryRoundTripper.retryConfig.ExceptOn) {
		for retries := 0; retries < maxReTries; retries++ {
			time.Sleep(retryRoundTripper.retryConfig.Delay)

			res, err := doRequest(req)
			if retries == maxReTries-1 {
				return res, err
			}

			if err != nil || !isExceptStatus(res.StatusCode, retryRoundTripper.retryConfig.ExceptOn) {
				continue
			}

			return res, nil
		}
	}

	return response, nil
}

func isExceptStatus(status int, exceptStatuses []int) bool {
	for _, exceptStatus := range exceptStatuses {
		if status == exceptStatus {
			return true
		}
	}
	return false
}

// RoundTripperMiddleware is a middleware which can be added as a transporter to a http client
// the middleware can be configured to retry the same request a number of times at certain intervals
func RoundTripperMiddleware(retryConfig Config) func(http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		return retryRoundTripper{
			retryConfig:  retryConfig,
			roundTripper: roundTripper,
		}
	}
}
