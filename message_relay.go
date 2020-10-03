package relay

import (
	"fmt"
	"sync"
)

// MessageRelayer defines a message relay which publishes messages to
// subscribers.
type MessageRelayer interface {
	SubscribeToMessages(msgType MessageType, messages chan<- Message)
	Start()
	Stop()
}

// NewMessageRelayer contructs a relayer
func NewMessageRelayer(socket NetworkSocket) *relayer {
	r := &relayer{
		socket:        socket,
		subscriptions: []Subscription{},
		queue:         NewQueue(),
		chMsgReceived: make(chan struct{}),
	}

	return r
}

// Subscription defines the MessageType to received and the channel on which to
// receive the message.
type Subscription struct {
	MsgType    MessageType
	MsgChannel chan<- Message
}

// relayer implements MesssageRelay
type relayer struct {
	mux           sync.Mutex
	socket        NetworkSocket
	subscriptions []Subscription
	chMsgReceived chan struct{}
	chStop        chan struct{}
	queue         *Queue

	verbose bool
}

func (r *relayer) SubscribeToMessages(msgType MessageType, messages chan<- Message) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.subscriptions = append(r.subscriptions, Subscription{
		MsgType:    msgType,
		MsgChannel: messages,
	})
}

func (r *relayer) Start() {
	r.chStop = make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go r.read(&wg)

	wg.Add(1)
	go r.broadcast(&wg)

	wg.Wait()
}

func (r *relayer) Stop() {
	close(r.chStop)
}

func (r *relayer) read(wg *sync.WaitGroup) {
	for {
		select {
		case <-r.chStop:
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
			if r.verbose {
				fmt.Printf("[Queued] Type=%d Data=%s\n", msg.Type, string(msg.Data))
			}
			r.chMsgReceived <- struct{}{}
		}
	}
}

func (r *relayer) broadcast(wg *sync.WaitGroup) {
	for {
		select {
		case <-r.chStop:
			wg.Done()

			return
		case <-r.chMsgReceived:
			msg := r.queue.Dequeue()

			if msg != nil {
				if r.verbose {
					fmt.Printf("[Broadcasting] Type=%d Data=%s\n", msg.Type, string(msg.Data))
				}

				for _, s := range r.subscriptions {
					// Check the subscriber wants to received the message type
					if s.MsgType == 0 || s.MsgType == msg.Type {
						// Send the message to the subscriber. Skip the broadcast if the
						// channel is blocked
						select {
						case s.MsgChannel <- *msg:
						default:
							if r.verbose {
								fmt.Println("Subscriber busy - Skipping", string(msg.Data))
							}
						}
					}
				}
			}
		}
	}
}

func (r *relayer) SetVerbose(verbose bool) {
	r.verbose = verbose
}
