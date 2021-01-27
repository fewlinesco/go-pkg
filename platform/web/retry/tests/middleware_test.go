package tests

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/fewlinesco/go-pkg/platform/web/retry"
)

func TestRetryRoundTripperMiddleware(t *testing.T) {
	type roundTripperMiddlewareTestCase struct {
		name              string
		expectedHTTPCode int
		httpCodesToReturn []int
		expectedCalls     int
		config            retry.Config
	}

	cfg := retry.Config{
		MaxRetry: 5,
		ExceptOn: []int{http.StatusOK, http.StatusNotFound},
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	httpClient := http.Client{
		Transport: retry.RoundTripperMiddleware(cfg)(transport),
	}

	tcs := []roundTripperMiddlewareTestCase{
		{
			name:              "it_does_not_retry_when_the_first_request_is_successful",
			expectedHTTPCode: 	http.StatusOK,
			httpCodesToReturn: []int{http.StatusOK},
			expectedCalls:     1,
		},
		{
			name:              "it_returns_the_last_response_when_no_request_is_successful",
			expectedHTTPCode: 	http.StatusForbidden,
			httpCodesToReturn: []int{
				http.StatusUnprocessableEntity,
				http.StatusUnprocessableEntity,
				http.StatusUnprocessableEntity,
				http.StatusUnprocessableEntity,
				http.StatusUnprocessableEntity,
				http.StatusForbidden,
			},
			expectedCalls:     6,
		},
		{
			name:              "it_stops_retrying_when_api_returns_expected_code",
			expectedHTTPCode: 	http.StatusOK,
			httpCodesToReturn: []int{http.StatusBadRequest, http.StatusBadRequest, http.StatusOK},
			expectedCalls:     3,
		},
		{
			name:              "it_stops_retrying_when_api_returns_other_expected_code",
			expectedHTTPCode: 	http.StatusNotFound,
			httpCodesToReturn: []int{
				http.StatusForbidden,
				http.StatusForbidden,
				http.StatusForbidden,
				http.StatusForbidden,
				http.StatusNotFound,
			},
			expectedCalls:     5,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
		tc := tc
			var requestTimestamps []time.Time

			handler := func() http.HandlerFunc {
				var lock sync.Mutex
				var count = 0

				return func(w http.ResponseWriter, r *http.Request) {
					defer func() {
						if err := recover(); err != nil {
							t.Fatalf("PANIC RECOVERED: %+v", err)
						}
					}()
					lock.Lock()
					defer lock.Unlock()

					if len(tc.httpCodesToReturn) <= count {
						t.Fatalf("the handler was called more than foreseen. Current count: %d, expectedCount: %d", count, len(tc.httpCodesToReturn))
					}

					requestTimestamps = append(requestTimestamps, time.Now())
					w.WriteHeader(tc.httpCodesToReturn[count])
					count++

				}
			}

			server := httptest.NewServer(handler())
			defer server.Close()

			res, err := httpClient.Get(server.URL)
			if err != nil {
				t.Fatalf("an error occured while dispatching the request: %v", err)
			}

			if len(requestTimestamps) != tc.expectedCalls {
				t.Fatalf("Expected the handler to be counted: %d times, but it was called: %d time(s)", tc.expectedCalls, len(requestTimestamps))
			}

			if res.StatusCode != tc.expectedHTTPCode {
				t.Fatalf("Expected the response to have the http code: %d, but it returned: %d", tc.expectedHTTPCode, res.StatusCode)
			}
		})
	}

}
