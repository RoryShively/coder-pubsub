package main

import (
	"testing"
	"time"
)

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
		for !sub.closed {
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
			var count int

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
							count++
						}
					}
				}()
			}

			// Send messages
			for i := 0; i < c.msgsSent; i++ {
				test_ds.StoreMessage("test_topic", []byte("testing123"))
			}

			// TODO: Set timeout timer and loop that expects correct result
			time.Sleep(3 * time.Second)

			expectedTotal := c.msgsSent * c.numSubs
			if count != expectedTotal {
				t.Errorf("Sent out %d messages to %d subscribers and received %d events, %d expected",
					c.msgsSent, c.numSubs, count, expectedTotal)
			}
		}
	})
}
