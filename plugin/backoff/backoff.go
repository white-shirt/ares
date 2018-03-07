/*
   github.com/cenkalti/backoff
*/
package backoff

import "time"

// BackOff is a backoff policy for retrying an operation.
type BackOff interface {
	// returns the duration to wait before retrying the operation
	Next() time.Duration
	// Reset to initial state.
	Reset()
}

// ZeroBackOff is a fixed backoff policy whose backoff time is always zero,
// meaning that the operation is retried immediately without waiting, indefinitely.
type ZeroBackOff struct{}

func (b *ZeroBackOff) Reset()              {}
func (b *ZeroBackOff) Next() time.Duration { return 0 }

// StopBackOff is a fixed backoff policy that always returns backoff.Stop for
// NextBackOff(), meaning that the operation should never be retried.
type StopBackOff struct{}

func (b *StopBackOff) Reset()              {}
func (b *StopBackOff) Next() time.Duration { return -1 }

// ConstantBackOff is a backoff policy that always returns the same backoff delay.
// This is in contrast to an exponential backoff policy,
// which returns a delay that grows longer as you call NextBackOff() over and over again.
type ConstantBackOff struct {
	Interval time.Duration
}

func (b *ConstantBackOff) Reset()              {}
func (b *ConstantBackOff) Next() time.Duration { return b.Interval }
