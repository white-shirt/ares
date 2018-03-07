package backoff

import (
	"testing"
	"time"
)

func TestNextBackOffMillis(t *testing.T) {
	subtestNextBackOff(t, 0, new(ZeroBackOff))
	subtestNextBackOff(t, Stop, new(StopBackOff))
}

func subtestNextBackOff(t *testing.T, expectedValue time.Duration, backOffPolicy BackOff) {
	for i := 0; i < 10; i++ {
		next := backOffPolicy.Next()
		if next != expectedValue {
			t.Errorf("got: %d expected: %d", next, expectedValue)
		}
	}
}

func TestConstantBackOff(t *testing.T) {
	backoff := new(ConstantBackOff)
	backoff.Interval = time.Second
	if backoff.Next() != time.Second {
		t.Error("invalid interval")
	}
}
