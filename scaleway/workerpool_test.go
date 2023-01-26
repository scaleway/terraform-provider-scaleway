package scaleway

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPoolSimple(t *testing.T) {
	pool := NewWorkerPool(2)

	pool.AddTask(func() error {
		return nil
	})

	pool.AddTask(func() error {
		return errors.New("error")
	})

	pool.AddTask(func() error {
		return nil
	})

	errs := pool.CloseAndWait()

	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "error", errs[0].Error())
}

func TestWorkerPoolWaitTime(t *testing.T) {
	pool := NewWorkerPool(2)

	pool.AddTask(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	pool.AddTask(func() error {
		time.Sleep(50 * time.Millisecond)
		return errors.New("error")
	})

	pool.AddTask(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	errs := pool.CloseAndWait()

	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "error", errs[0].Error())
}

func TestWorkerPoolWaitTimeMultiple(t *testing.T) {
	pool := NewWorkerPool(5)
	iterations := 20

	for i := 0; i < iterations; i++ {
		copyOfI := i

		pool.AddTask(func() error {
			time.Sleep(100 * time.Millisecond)

			if copyOfI%2 == 0 {
				return fmt.Errorf("error %d", copyOfI)
			}

			return nil
		})
	}

	errs := pool.CloseAndWait()

	assert.Equal(t, iterations/2, len(errs))

	for i := 0; i < iterations; i++ {
		if i%2 == 0 {
			found := false
			for _, err := range errs {
				if err.Error() == fmt.Sprintf("error %d", i) {
					found = true
					break
				}
			}

			assert.True(t, found)
		}
	}
}
