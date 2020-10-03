package relay

import (
	"errors"
	"sync"
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

// MockSubscriber mocks a subscriber and decrements a wait group counter when
// the expected number of Process calls have been made.
type MockSubscriber struct {
	mock.Mock
	mux       sync.Mutex
	callCount int
	doneCount int
	wg        *sync.WaitGroup
}

func NewMockSubscriber(doneCount int, wg *sync.WaitGroup) *MockSubscriber {
	return &MockSubscriber{
		callCount: 0,
		doneCount: doneCount,
		wg:        wg,
	}
}

// Process mocks a subscriber Process call and decrements a WaitGroup counter
func (ms *MockSubscriber) Process(msg Message) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()

	ms.Called(msg)

	ms.callCount++
	if ms.callCount == ms.doneCount {
		ms.wg.Done()
	}

	return nil
}

func TestSimpleStartNewRoundQueue(t *testing.T) {
	var wg sync.WaitGroup

	mockSocket := new(MockNetworkSocket)

	mockSocket.On("Read").Return(NewMessage(StartNewRound, "1"), nil).After(500 * time.Millisecond).Once()
	mockSocket.On("Read").Return(NewMessage(StartNewRound, "2"), nil).After(500 * time.Millisecond).Once()
	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

	mockSubscriber1 := NewMockSubscriber(2, &wg)
	wg.Add(1)
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "1"))
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "2"))

	relayer := NewMessageRelayer(mockSocket)

	ch1 := make(chan Message)

	go func() {
		for {
			msg := <-ch1

			mockSubscriber1.Process(msg)
		}
	}()

	// Subscriber1 reads all messages
	relayer.SubscribeToMessages(0, ch1)

	go relayer.Start()

	wg.Wait()

	relayer.Stop()

	mockSubscriber1.AssertNumberOfCalls(t, "Process", 2)
}

func TestSimpleReceivedAnswerQueue(t *testing.T) {
	var wg sync.WaitGroup

	mockSocket := new(MockNetworkSocket)

	mockSocket.On("Read").Return(NewMessage(ReceivedAnswer, "1"), nil).After(50 * time.Millisecond).Once()
	mockSocket.On("Read").Return(NewMessage(ReceivedAnswer, "2"), nil).After(50 * time.Millisecond).Once()
	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

	mockSubscriber1 := NewMockSubscriber(2, &wg)
	wg.Add(1)

	mockSubscriber1.On("Process", NewMessage(ReceivedAnswer, "1"))
	mockSubscriber1.On("Process", NewMessage(ReceivedAnswer, "2"))

	relayer := NewMessageRelayer(mockSocket)

	ch1 := make(chan Message)

	go func() {
		for {
			msg := <-ch1

			mockSubscriber1.Process(msg)
		}
	}()

	// Subscriber1 reads all messages
	relayer.SubscribeToMessages(0, ch1)

	go relayer.Start()

	wg.Wait()

	relayer.Stop()

	mockSubscriber1.AssertNumberOfCalls(t, "Process", 2)
}

func TestSimpleAlternatingMessages(t *testing.T) {
	var wg sync.WaitGroup
	mockSocket := new(MockNetworkSocket)

	for i := 0; i <= 3; i++ {
		mockSocket.On("Read").
			Return(NewMessage(MessageType(i%2+1), "data"), nil).
			After(50 * time.Millisecond).
			Once()
	}
	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

	mockSubscriber1 := NewMockSubscriber(4, &wg)
	wg.Add(1)
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "data")).Times(2)
	mockSubscriber1.On("Process", NewMessage(ReceivedAnswer, "data")).Times(2)
	mockSubscriber2 := NewMockSubscriber(2, &wg)
	wg.Add(1)
	mockSubscriber2.On("Process", NewMessage(StartNewRound, "data")).Times(2)
	mockSubscriber3 := NewMockSubscriber(2, &wg)
	wg.Add(1)
	mockSubscriber3.On("Process", NewMessage(ReceivedAnswer, "data")).Times(2)

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

	wg.Wait()

	relayer.Stop()

	mockSubscriber1.AssertNumberOfCalls(t, "Process", 4)
	mockSubscriber2.AssertNumberOfCalls(t, "Process", 2)
	mockSubscriber3.AssertNumberOfCalls(t, "Process", 2)
}

func TestBusySubscriber(t *testing.T) {
	var wg sync.WaitGroup
	mockSocket := new(MockNetworkSocket)

	mockSocket.On("Read").Return(NewMessage(StartNewRound, "1"), nil).Once()
	// Skip this message. The read time is set to be during the processing time of the subscriber
	mockSocket.On("Read").Return(NewMessage(StartNewRound, "2"), nil).After(25 * time.Millisecond).Once()
	mockSocket.On("Read").Return(NewMessage(StartNewRound, "3"), nil).After(75 * time.Millisecond).Once()
	mockSocket.On("Read").Return(Message{}, errors.New("No more data"))

	mockSubscriber1 := NewMockSubscriber(2, &wg)
	wg.Add(1)
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "1"))
	mockSubscriber1.On("Process", NewMessage(StartNewRound, "3"))

	relayer := NewMessageRelayer(mockSocket)

	ch1 := make(chan Message)

	go func() {
		for {
			msg := <-ch1

			time.Sleep(50 * time.Millisecond)

			mockSubscriber1.Process(msg)
		}
	}()

	// Subscriber1 reads all messages
	relayer.SubscribeToMessages(0, ch1)

	go relayer.Start()

	wg.Wait()

	relayer.Stop()

	mockSubscriber1.AssertNumberOfCalls(t, "Process", 2)
}

// These would only test the priority handling of the queue, which is tested in
// the queue tests itself.
// func TestStartNewRoundBroadcastFirst(t *testing.T) {
// }

// func TestMostRecentReceivedAnswer(t *testing.T) {
// }

// func TestMostRecentStartNewRounds(t *testing.T) {
// }
