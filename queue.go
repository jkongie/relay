package relay

import (
	"sync"
)

// StartNewRoundQueueLimit defines the number of messages that can be queued for
// the StartNewRound message type
const StartNewRoundQueueLimit = 2

// ReceivedAnswerQueueLimit defines the number of messages that can be queued for
// the ReceivedAnswer message type
const ReceivedAnswerQueueLimit = 1

// Queue stores the messages that have been read in by the relay. It takes a
// unique queue implemntation because we must ensure that the most recent
// StartNewRound and ReceivedAnswer messages are delivered. As such,
// ReceivedAnswer messages are stored separately from the StartNewRound messages
// queue.
type Queue struct {
	mux sync.Mutex
	// Contains the StartNewRound messages
	startMessages []Message
	// Contains the ReceivedAnswer messages
	receivedMessages []Message
}

// NewQueue constructs a new empty queue
func NewQueue() *Queue {
	return &Queue{
		startMessages:    []Message{},
		receivedMessages: []Message{},
	}
}

// Enqueue pushes the message onto the queue.
func (q *Queue) Enqueue(msg Message) {
	q.mux.Lock()
	defer q.mux.Unlock()

	switch msg.Type {
	case StartNewRound:
		// Remove the oldest message
		if len(q.startMessages) >= StartNewRoundQueueLimit {
			q.startMessages = q.startMessages[1:]
		}

		// Append the latest message
		q.startMessages = append(q.startMessages, msg)
	case ReceivedAnswer:
		// Remove the oldest message
		if len(q.receivedMessages) >= ReceivedAnswerQueueLimit {
			q.receivedMessages = q.receivedMessages[1:]
		}

		// Append the latest message
		q.receivedMessages = append(q.receivedMessages, msg)
	}
}

// Dequeue the message from the queue. StartNewRound messages are given dequeue
// priority.
func (q *Queue) Dequeue() *Message {
	q.mux.Lock()
	defer q.mux.Unlock()

	// Prioritise any StartNewRound messages in the queue
	if len(q.startMessages) > 0 {
		// Pop the msg
		msg := q.startMessages[len(q.startMessages)-1]
		q.startMessages = q.startMessages[:len(q.startMessages)-1]

		return &msg
	}

	if len(q.receivedMessages) > 0 {
		// Pop the msg
		msg := q.receivedMessages[len(q.receivedMessages)-1]
		q.receivedMessages = q.receivedMessages[:len(q.receivedMessages)-1]

		return &msg
	}

	return nil
}
