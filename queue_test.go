package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	var tests = []struct {
		input    []Message
		expected []Message
	}{
		// Receiving only ReceivedAnswer
		{[]Message{
			NewMessage(ReceivedAnswer, "1"),
			NewMessage(ReceivedAnswer, "2"),
			NewMessage(ReceivedAnswer, "3"),
		}, []Message{
			NewMessage(ReceivedAnswer, "3"),
		}},
		// Receiving only StartNewRound
		{[]Message{
			NewMessage(StartNewRound, "1"),
			NewMessage(StartNewRound, "2"),
			NewMessage(StartNewRound, "3"),
			NewMessage(StartNewRound, "4"),
		}, []Message{
			NewMessage(StartNewRound, "4"),
			NewMessage(StartNewRound, "3"),
		}},
		// Receiving a mix of types
		{[]Message{
			NewMessage(StartNewRound, "1"),
			NewMessage(StartNewRound, "2"),
			NewMessage(ReceivedAnswer, "1"),
			NewMessage(StartNewRound, "4"),
			NewMessage(StartNewRound, "5"),
			NewMessage(ReceivedAnswer, "2"),
		}, []Message{
			NewMessage(StartNewRound, "5"),
			NewMessage(StartNewRound, "4"),
			NewMessage(ReceivedAnswer, "2"),
		}},
	}

	for _, test := range tests {
		queue := NewQueue()

		for _, msg := range test.input {
			queue.Enqueue(msg)
		}

		for _, exp := range test.expected {
			assert.Equal(t, *queue.Dequeue(), exp)
		}
		// Ensure the queue is empty
		assert.Nil(t, queue.Dequeue())
	}
}
