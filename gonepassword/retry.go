package gonepassword

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type nonRetryableError struct {
	Message string
}

func (e nonRetryableError) Error() string {
	return e.Message
}

type retryAbleFunc func() (any, error)
type backOffFunc func(attempt int) time.Duration

func exponentialBackoff(attempt int) time.Duration {
	return time.Duration(2<<uint(attempt)) * time.Second
}

func retry(retries int, backoff backOffFunc, f retryAbleFunc) (any, error) {
	var output any
	var err error
	var nonRetryableError *nonRetryableError
	for i := 0; i < retries; i++ {
		output, err = f()
		if err == nil {
			break
		}
		if errors.As(err, &nonRetryableError) {
			break
		}
		if i <= retries {
			backoffTime := backoff(i)
			fmt.Fprintf(os.Stderr, "retrying in %.0f seconds...\n", backoffTime.Seconds())
			time.Sleep(backoffTime)
		}
	}
	return output, err
}
