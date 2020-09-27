package relay

import (
	"fmt"
	"sync"
)

type MessageType int

const (
	StartNewRound MessageType = 1 << iota
	ReceivedAnswer
)

type NetworkSocket interface {
	Read() (Message, error)
}

// Message represents a message received from the network
type Message struct {
	Type MessageType
	Data []byte
}

// NewMessage constructs a new message
func NewMessage(t MessageType, data string) Message {
	return Message{
		Type: t,
		Data: []byte(data),
	}
}

type MessageRelayer interface {
	SubscribeToMessages(msgType MessageType, messages chan<- Message)
	Start()
	Stop()
}

func NewMessageRelayer(socket NetworkSocket) *relayer {
	r := &relayer{
		socket:        socket,
		subscriptions: []Subscription{},
		queue:         NewQueue(),
		readDone:      make(chan bool),
		broadcastDone: make(chan bool),
	}

	return r
}

type Subscription struct {
	MsgType    MessageType
	MsgChannel chan<- Message
}

//
type relayer struct {
	socket        NetworkSocket
	subscriptions []Subscription
	queue         *Queue
	readDone      chan bool
	broadcastDone chan bool
}

func (r *relayer) SubscribeToMessages(msgType MessageType, messages chan<- Message) {
	r.subscriptions = append(r.subscriptions, Subscription{
		MsgType:    msgType,
		MsgChannel: messages,
	})
}

func (r *relayer) Start() {
	var wg sync.WaitGroup

	wg.Add(1)
	go r.read(&wg)
	wg.Add(1)
	go r.broadcast(&wg)

	wg.Wait()
}

func (r *relayer) Shutdown() {
	r.readDone <- true
	r.broadcastDone <- true
}

func (r *relayer) read(wg *sync.WaitGroup) {
	for {
		select {
		case <-r.readDone:
			wg.Done()

			return
		default:
			msg, err := r.socket.Read()

			// Continue if an error occurs. This could quit depending on the
			// error returned.
			if err != nil {
				continue
			}

			// Enqueue the message
			r.queue.Enqueue(msg)
		}
	}
}

func (r *relayer) broadcast(wg *sync.WaitGroup) {
	shutdown := false

	for {
		select {
		case shutdown = <-r.broadcastDone:
		default:
			if len(r.subscriptions) > 0 {
				msg := r.queue.Dequeue()

				if msg != nil {
					fmt.Printf("[Broadcasting] Type=%d Data=%s\n", msg.Type, string(msg.Data))

					for _, s := range r.subscriptions {
						// Check the subscriber wants to received the message type
						if s.MsgType == 0 || s.MsgType == msg.Type {
							// Send the message to the subscriber. Skip the broadcast if the
							// channel is blocked
							select {
							case s.MsgChannel <- *msg:
							default:
								fmt.Println("no message sent", string(msg.Data))
							}
						}
					}
				} else {
					if shutdown {
						wg.Done()

						return
					}
				}
			}
		}
	}
}
