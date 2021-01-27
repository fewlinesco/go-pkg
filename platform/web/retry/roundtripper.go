package retry

import (
	"net/http"
	"time"
)

type Config struct {
	MaxRetry int
	Delay    time.Duration
	ExceptOn []int
}

type retryRoundTripper struct {
	roundTripper http.RoundTripper
	retryConfig  Config
}

func (retryRoundTripper retryRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	try := retryRoundTripper.retryConfig.MaxRetry + 1
	for try > 0 {
		res, err = retryRoundTripper.roundTripper.RoundTrip(req)
		if err != nil {
			return res, err
		}

		if res.StatusCode >= 200 && res.StatusCode < 300 {
			return res, nil
		}

		if isExceptStatus(res.StatusCode, retryRoundTripper.retryConfig.ExceptOn) {
			return res, nil
		}

		try = try - 1
		time.Sleep(retryRoundTripper.retryConfig.Delay)
	}
	return res, err
}

func isExceptStatus(status int, exceptStatuses []int) bool {
	for exceptStatus := range exceptStatuses {
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
