/*
Interesting reference on backoff:
- Exponential Backoff And Jitter (AWS Blog):
  https://www.awsarchitectureblog.com/2015/03/backoff.html

We use Jitter as a default for exponential backoff, as the goal of
this module is not to provide precise 'ticks', but good behaviour to
implement retries that are helping the server to recover faster in
case of congestion.

It can be used in several ways:
- Using duration to get next sleep time.
- Using ticker channel to trigger callback function on tick

The functions for Backoff are not threadsafe, but you can:
- Keep the attempt counter on your end and use durationForAttempt(int)
- Use lock in your own code to protect the Backoff structure.

TODO: Implement Backoff Ticker channel
TODO: Implement throttler interface. Throttler could be used to implement various reconnect strategies.
*/

package xmpp

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultBase   int = 20 // Backoff base, in ms
	defaultFactor int = 2
	defaultCap    int = 180000 // 3 minutes
)

// backoff provides increasing duration with the number of attempt
// performed. The structure is used to support exponential backoff on
// connection attempts to avoid hammering the server we are connecting
// to.
type backoff struct {
	NoJitter     bool
	Base         int
	Factor       int
	Cap          int
	lastDuration int
	attempt      int
}

// duration returns the duration to apply to the current attempt.
func (b *backoff) duration() time.Duration {
	d := b.durationForAttempt(b.attempt)
	b.attempt++
	return d
}

// wait sleeps for backoff duration for current attempt.
func (b *backoff) wait() {
	time.Sleep(b.duration())
}

// durationForAttempt returns a duration for an attempt number, in a stateless way.
func (b *backoff) durationForAttempt(attempt int) time.Duration {
	b.setDefault()
	expBackoff := math.Min(float64(b.Cap), float64(b.Base)*math.Pow(float64(b.Factor), float64(b.attempt)))
	d := int(math.Trunc(expBackoff))
	if !b.NoJitter {
		d = rand.Intn(d)
	}
	return time.Duration(d) * time.Millisecond
}

// reset sets back the number of attempts to 0. This is to be called after a successful operation has been performed,
// to reset the exponential backoff interval.
func (b *backoff) reset() {
	b.attempt = 0
}

func (b *backoff) setDefault() {
	if b.Base == 0 {
		b.Base = defaultBase
	}

	if b.Cap == 0 {
		b.Cap = defaultCap
	}

	if b.Factor == 0 {
		b.Factor = defaultFactor
	}
}

/*
We use full jitter as default for now as it seems to provide good behaviour for reconnect.

Base is the default interval between attempts (if backoff Factor was equal to 1)

Attempt is the number of retry for operation. If we start attempt at 0, first sleep equals base.

Cap is the maximum sleep time duration we tolerate between attempts
*/
