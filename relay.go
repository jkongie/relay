package relay

// MessageType defines the type of message to be processed
type MessageType int

const (
	// StartNewRound defines a StartNewRound MessageType
	StartNewRound MessageType = 1 << iota
	// ReceivedAnswer defines a ReceivedAnswer MessageType
	ReceivedAnswer
)

// NetworkSocket defines a socket which reads messsages to the caller.
type NetworkSocket interface {
	Read() (Message, error)
}

// Message represents a message received from the network socket
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
