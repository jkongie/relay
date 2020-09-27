package relay

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockNetworkSocket struct {
	mock.Mock
}

func (skt *MockNetworkSocket) Read() (Message, error) {
	args := skt.Called()

	return args.Get(0).(Message), args.Error(1)
}

type MockSubscriber struct {
	mock.Mock
}

func (ms *MockSubscriber) Process(msg Message) error {
	ms.Called(msg)

	return nil
}

func TestSimpleAlternatingMessages(t *testing.T) {
	mockSocket := new(MockNetworkSocket)

	for i := 0; i <= 1; i++ {
		mockSocket.On("Read").Return(NewMessage(MessageType(i%2+1), "data"), nil).After(500 * time.Millisecond).Once()
	}
	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

	mockSubscriber1 := new(MockSubscriber)
	mockSubscriber1.On("Process", mock.AnythingOfType("Message"))
	mockSubscriber2 := new(MockSubscriber)
	mockSubscriber2.On("Process", mock.AnythingOfType("Message"))
	mockSubscriber3 := new(MockSubscriber)
	mockSubscriber3.On("Process", mock.AnythingOfType("Message"))

	relayer := NewMessageRelayer(mockSocket)

	ch1 := make(chan Message)
	ch2 := make(chan Message)
	ch3 := make(chan Message)

	go func() {
		for {
			msg := <-ch1

			mockSubscriber1.Process(msg)
		}
	}()

	go func() {
		for {
			msg := <-ch2

			mockSubscriber2.Process(msg)
		}
	}()

	go func() {
		for {
			msg := <-ch3

			mockSubscriber3.Process(msg)
		}
	}()

	// Subscriber1 reads all messages
	relayer.SubscribeToMessages(0, ch1)
	// Subscriber2 reads all StartNewRound messages
	relayer.SubscribeToMessages(StartNewRound, ch2)
	// Subscriber2 reads all ReceivedAnswer messages
	relayer.SubscribeToMessages(ReceivedAnswer, ch3)

	go relayer.Start()

	// Sleeping here to give time for the test to run. Not sure how else to handle this.
	time.Sleep(2 * time.Second)

	relayer.Shutdown()

	mockSubscriber1.AssertNumberOfCalls(t, "Process", 2)
	mockSubscriber2.AssertNumberOfCalls(t, "Process", 1)
	mockSubscriber3.AssertNumberOfCalls(t, "Process", 1)
}

func TestBusySubscriber(t *testing.T) {
	mockSocket := new(MockNetworkSocket)

	mockSocket.On("Read").Return(NewMessage(StartNewRound, "1"), nil).After(500 * time.Millisecond).Once()
	mockSocket.On("Read").Return(NewMessage(StartNewRound, "2"), nil).Once()
	mockSocket.On("Read").Return(NewMessage(StartNewRound, "3"), nil).After(600 * time.Millisecond).Once()
	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

	mockSubscriber1 := new(MockSubscriber)
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "1"))
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "3"))

	relayer := NewMessageRelayer(mockSocket)

	ch1 := make(chan Message)

	go func() {
		for {
			msg := <-ch1

			time.Sleep(1000)

			mockSubscriber1.Process(msg)
		}
	}()

	// Subscriber1 reads all messages
	relayer.SubscribeToMessages(0, ch1)

	go relayer.Start()

	// Sleeping here to give time for the test to run. Not sure how else to handle this.
	time.Sleep(2 * time.Second)

	relayer.Shutdown()

	mockSubscriber1.AssertNumberOfCalls(t, "Process", 2)
}

// func TestStartNewRoundBroadcastFirst(t *testing.T) {
// 	fmt.Println("TestStartNewRoundBroadcastFirst")
// 	mockSocket := new(MockNetworkSocket)

// 	mockSocket.On("Read").Return(NewMessage(StartNewRound, "snr1"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(ReceivedAnswer, "ra1"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(StartNewRound, "snr2"), nil).Once()
// 	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

// 	mockSubscriber1 := new(MockSubscriber)
// 	mockSubscriber1.On("Process", NewMessage(StartNewRound, "snr1"))
// 	mockSubscriber1.On("Process", NewMessage(StartNewRound, "snr2"))
// 	mockSubscriber1.On("Process", NewMessage(ReceivedAnswer, "ra1"))

// 	relayer := NewMessageRelayer(mockSocket)

// 	ch1 := make(chan Message)

// 	go func() {
// 		for {
// 			msg := <-ch1
// 			fmt.Printf("\t[Subscriber 1] Processed Type=%d Data=%s\n", msg.Type, string(msg.Data))

// 			// time.Sleep(1000)

// 			mockSubscriber1.Process(msg)
// 		}
// 	}()

// 	// Subscriber1 reads all messages
// 	relayer.SubscribeToMessages(0, ch1)

// 	go relayer.Start()

// 	// Sleeping here to give time for the test to run. Not sure how else to handle this.
// 	time.Sleep(3 * time.Second)

// 	relayer.Shutdown()

// 	mockSubscriber1.AssertNumberOfCalls(t, "Process", 3)
// }

// func TestMostRecentReceivedAnswer(t *testing.T) {
// 	mockSocket := new(MockNetworkSocket)

// 	mockSocket.On("Read").Return(NewMessage(ReceivedAnswer, "1"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(ReceivedAnswer, "2"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(ReceivedAnswer, "3"), nil).Once()

// 	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

// 	mockSubscriber1 := new(MockSubscriber)
// 	mockSubscriber1.On("Process", NewMessage(ReceivedAnswer, "1"))
// 	mockSubscriber1.On("Process", NewMessage(ReceivedAnswer, "3"))

// 	relayer := NewMessageRelayer(mockSocket)

// 	ch1 := make(chan Message)

// 	go func() {
// 		for {
// 			msg := <-ch1
// 			fmt.Printf("\t[Subscriber 1] Processed Type=%d Data=%s\n", msg.Type, string(msg.Data))

// 			// time.Sleep(1000)

// 			mockSubscriber1.Process(msg)
// 		}
// 	}()

// 	// Subscriber1 reads all messages
// 	relayer.SubscribeToMessages(0, ch1)

// 	go relayer.Start()

// 	// Sleeping here to give time for the test to run. Not sure how else to handle this.
// 	time.Sleep(3 * time.Second)

// 	relayer.Shutdown()

// 	mockSubscriber1.AssertNumberOfCalls(t, "Process", 2)
// }

// func TestMostRecentStartNewRounds(t *testing.T) {
// 	mockSocket := new(MockNetworkSocket)

// 	mockSocket.On("Read").Return(NewMessage(StartNewRound, "1"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(StartNewRound, "2"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(StartNewRound, "3"), nil).Once()
// 	mockSocket.On("Read").Return(NewMessage(StartNewRound, "4"), nil).Once()

// 	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

// 	mockSubscriber1 := new(MockSubscriber)
// 	mockSubscriber1.On("Process", NewMessage(StartNewRound, "1"))
// 	mockSubscriber1.On("Process", NewMessage(StartNewRound, "3"))
// 	mockSubscriber1.On("Process", NewMessage(StartNewRound, "4"))

// 	relayer := NewMessageRelayer(mockSocket)

// 	ch1 := make(chan Message)

// 	go func() {
// 		for {
// 			msg := <-ch1
// 			fmt.Printf("\t[Subscriber 1] Processed Type=%d Data=%s\n", msg.Type, string(msg.Data))

// 			// time.Sleep(1000)

// 			mockSubscriber1.Process(msg)
// 		}
// 	}()

// 	// Subscriber1 reads all messages
// 	relayer.SubscribeToMessages(0, ch1)

// 	go relayer.Start()

// 	// Sleeping here to give time for the test to run. Not sure how else to handle this.
// 	time.Sleep(3 * time.Second)

// 	relayer.Shutdown()

// 	mockSubscriber1.AssertNumberOfCalls(t, "Process", 3)
// }
