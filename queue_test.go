package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
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
			NewMessage(StartNewRound, "3"),
			NewMessage(StartNewRound, "4"),
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
			NewMessage(ReceivedAnswer, "2"),
			NewMessage(StartNewRound, "4"),
			NewMessage(StartNewRound, "5"),
		}},
	}

	for _, test := range tests {
		queue := NewQueue()

		for _, message := range test.input {
			queue.Enqueue(message)
		}

		assert.Equal(t, queue.GetMessages(), test.expected)
	}
}

func TestDequeue(t *testing.T) {
	queue := &Queue{
		messages: []Message{
			NewMessage(StartNewRound, "1"),
			NewMessage(StartNewRound, "2"),
			NewMessage(StartNewRound, "3"),
		},
	}

	assert.Equal(t, *queue.Dequeue(), NewMessage(StartNewRound, "3"))
	assert.Equal(t, *queue.Dequeue(), NewMessage(StartNewRound, "2"))
	assert.Equal(t, *queue.Dequeue(), NewMessage(StartNewRound, "1"))
	assert.Nil(t, queue.Dequeue())
}
