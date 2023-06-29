package utils

import "time"

func CallWithRetries[T any](fn func() (T, error), retryCount int) (result T, err error) {
	for i := 0; i < retryCount; i++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		time.Sleep(3 * time.Second)
	}
	return result, err
}
