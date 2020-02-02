package main

import (
	"sync/atomic"
	"testing"
	"time"
)

// Utility type to help count concurrently.
// uint used because it's a counter
type count32 uint32

func (c *count32) increment() uint32 {
	return atomic.AddUint32((*uint32)(c), 1)
}

func (c *count32) get() uint32 {
	return atomic.LoadUint32((*uint32)(c))
}

func TestDatastore(t *testing.T) {
	t.Run("subscription addition/removal", func(t *testing.T) {
		test_ds := NewDataStore()
		_, sub, err := test_ds.RetrieveWithSubscription("test_topic")
		if err != nil {
			t.Error(err)
		}

		// Realive channels so they done block
		go func() {
			for {
				select {
				case _ = <-sub.msgs:
					continue
				}
			}
		}()

		// Check sub was added
		if len(test_ds.subscribers) != 1 {
			t.Error("subscriber was never added")
		}

		// Close sub and start removal by calling notify.
		// Open loop while waiting for channel to close.
		close(sub.done)
		for !sub.IsClosed() {
		}
		test_ds.notifySubscribers(MessageEvent{})

		// Check sub was removed
		if len(test_ds.subscribers) != 0 {
			t.Error("subscriber was never removed")
		}
	})

	t.Run("subscription event sent", func(t *testing.T) {
		cases := []struct {
			numSubs  int
			msgsSent int
		}{
			{
				10,
				10,
			},
			{
				0, // Edgecase: no subs
				19,
			},
			{
				3,
				20,
			},
		}
		for _, c := range cases {
			test_ds := NewDataStore()
			var count count32

			// Create subs
			for i := 0; i < c.numSubs; i++ {
				_, sub, err := test_ds.RetrieveWithSubscription("test_topic")
				if err != nil {
					t.Error(err)
				}

				// Count received msgs
				go func() {
					for {
						select {
						case _ = <-sub.msgs:
							count.increment()
						}
					}
				}()
			}

			// Send messages
			for i := 0; i < c.msgsSent; i++ {
				test_ds.StoreMessage("test_topic", []byte("testing123"))
			}

			// Set timeout so loop doesn't pause forever waiting on
			// correct result
			var timedOut bool
			timer := time.NewTimer(time.Second)
			go func() {
				<-timer.C
				timedOut = true
			}()

			// Test fails if counts don't add up when timeout occurs
			expectedTotal := uint32(c.msgsSent * c.numSubs)
			for {
				success := count.get() == expectedTotal
				if !success && timedOut {
					t.Errorf("Sent out %d messages to %d subscribers and received %d events, %d expected",
						c.msgsSent, c.numSubs, count, expectedTotal)
					break
				} else if success {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}

		}
	})
}
