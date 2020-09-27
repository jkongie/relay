package relay

import (
	"sync"
)

// Queue stores the messages that have been read in by the relay. It takes a
// unique queue implemntation because we must ensure that the most recent
// StartNewRound and ReceivedAnswer messages are delivered. As such,
// ReceivedAnswer messages are stored separately from the StartNewRound messages
// queue.
type Queue struct {
	mux      sync.Mutex
	messages []Message
}

// NewQueue constructs a new empty queue
func NewQueue() *Queue {
	return &Queue{
		messages: []Message{},
	}
}

// Enqueue pushes the message onto the queue.
func (q *Queue) Enqueue(msg Message) {
	q.mux.Lock()
	defer q.mux.Unlock()

	switch msg.Type {
	case StartNewRound:
		count := 0
		for _, msg := range q.messages {
			if msg.Type == StartNewRound {
				count++
			}
		}

		// Delete the oldest StartNewRound message when the limit is reached
		if len(q.messages) > 0 && count >= 2 {
			idx := 0
			if q.messages[0].Type == ReceivedAnswer {
				idx = 1
			}

			q.messages = append(q.messages[:idx], q.messages[idx+1:]...)
		}

		// Append the message
		q.messages = append(q.messages, msg)
	case ReceivedAnswer:
		// Replace the most recent received answer otherwise unshift it into
		// the queue. We know that the ReceivedAnswer message is always the
		// first item in the slice because we unshift any addition of that type
		// of message
		if len(q.messages) > 0 && q.messages[0].Type == ReceivedAnswer {
			q.messages[0] = msg
		} else {
			q.messages = append([]Message{msg}, q.messages...)
		}
	}
}

// Dequeue the message from the queue. StartNewRound messages are given dequeue
// priority.
func (q *Queue) Dequeue() *Message {
	q.mux.Lock()
	defer q.mux.Unlock()

	if len(q.messages) > 0 {
		msg := q.messages[len(q.messages)-1]
		q.messages = q.messages[:len(q.messages)-1]

		return &msg
	}

	return nil
}

// GetMessages is an accessor to messages
func (q *Queue) GetMessages() []Message {
	return q.messages
}
