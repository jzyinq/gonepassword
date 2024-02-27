package gonepassword

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func MilliExponentialBackoff(attempt int) time.Duration {
	// that will print 0 seconds in output but hey - it will be fast for tests
	return time.Duration(2<<uint(attempt)) * time.Millisecond
}

func TestRetry(t *testing.T) { //nolint
	recoveredRetry := 0
	testCases := []struct {
		name           string
		retries        int
		backoff        backOffFunc
		f              retryAbleFunc
		expectedOutput any
		expectedError  string
		expectedStderr string
	}{
		{
			name:           "should return output if no error",
			retries:        3,
			backoff:        MilliExponentialBackoff,
			f:              func() (any, error) { return "output", nil },
			expectedOutput: "output",
			expectedError:  "",
			expectedStderr: "",
		},
		{
			name:           "should retry on error",
			retries:        3,
			backoff:        MilliExponentialBackoff,
			f:              func() (any, error) { return nil, errors.New("error") },
			expectedOutput: nil,
			expectedError:  "error",
			expectedStderr: "retrying in 0 seconds...\nretrying in 0 seconds...\nretrying in 0 seconds...\n",
		},
		{
			name:    "should return output on successful retry",
			retries: 3,
			backoff: MilliExponentialBackoff,
			f: func() (any, error) {
				recoveredRetry++
				if recoveredRetry > 2 {
					return "success", nil
				}
				return nil, errors.New("error")
			},
			expectedOutput: "success",
			expectedError:  "",
			expectedStderr: "retrying in 0 seconds...\nretrying in 0 seconds...\n",
		},
		{
			name:    "should return output if error is non-retryable",
			retries: 3,
			backoff: MilliExponentialBackoff,
			f: func() (any, error) {
				return "output", &nonRetryableError{Message: "non-retryable error"}
			},
			expectedOutput: "output",
			expectedError:  "non-retryable error",
			expectedStderr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capturedStderr, output, err := captureStderrAndCallFunc(func() (any, error) {
				return retry(tc.retries, tc.backoff, tc.f)
			})

			if tc.expectedError != "" && (err == nil || err.Error() != tc.expectedError) {
				t.Errorf("Expected error %q, got %q", tc.expectedError, err)
			}

			if tc.expectedOutput != output {
				t.Errorf("Expected output %v, got %v", tc.expectedOutput, output)
			}

			if tc.expectedStderr != "" && !strings.Contains(capturedStderr, tc.expectedStderr) {
				t.Errorf("Expected stderr to contain %q, got %q", tc.expectedStderr, capturedStderr)
			}
		})
	}
}

func captureStderrAndCallFunc(f func() (any, error)) (capturedStderr string, output any, err error) {
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			return
		}
		outC <- buf.String()
	}()

	output, err = f()

	w.Close() //nolint

	os.Stderr = originalStderr
	capturedStderr = <-outC
	return capturedStderr, output, err
}
