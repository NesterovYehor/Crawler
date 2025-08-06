package queue

import (
	"sync"
	"time"
)

const fixedQueueBackoff = 1 * time.Second

var (
	HighPriorityQueue   = "queue:fetch:high"
	MediumPriorityQueue = "queue:fetch:medium"
	RetryPriorityQueue  = "queue:fetch:retry"
	StoreQueue          = "queue:store"
)

var FallbackOrder = []string{
	HighPriorityQueue,
	MediumPriorityQueue,
	StoreQueue,
	RetryPriorityQueue,
}

type QueueState struct {
	Name                string
	NextAttemptDue      time.Time
	ConsecutiveFailures int
}

type Source struct {
	mu   *sync.Mutex
	Curr string
	init string

	queueStates map[string]*QueueState
}

func NewSource(initialQueueName string) *Source {
	s := &Source{
		mu:          &sync.Mutex{},
		Curr:        initialQueueName,
		init:        initialQueueName,
		queueStates: make(map[string]*QueueState),
	}
	for _, qName := range FallbackOrder {
		s.queueStates[qName] = &QueueState{
			Name:           qName,
			NextAttemptDue: time.Now(),
		}
	}
	return s
}

func (s *Source) MarkQueueFailed() {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.queueStates[s.Curr]
	if state == nil {
		return
	}

	state.ConsecutiveFailures++
	state.NextAttemptDue = time.Now().Add(fixedQueueBackoff)
}

func (s *Source) GetCurrentSource() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.checkInit() {
		return s.Curr
	}

	currState := s.queueStates[s.Curr]
	if currState != nil && time.Now().After(currState.NextAttemptDue) {
		return s.Curr
	}
	s.findNewCurrSource()
	return s.Curr
}

func (s *Source) checkInit() bool {
	initialState := s.queueStates[s.init]
	initialAvailable := (initialState != nil && time.Now().After(initialState.NextAttemptDue))
	if initialAvailable && s.Curr != s.init {
		s.Curr = s.init
		return true
	}
	return false
}

func (s *Source) findNewCurrSource() {
	for _, qName := range FallbackOrder {
		if qName == s.Curr {
			continue
		}
		state := s.queueStates[qName]
		if state != nil && time.Now().After(state.NextAttemptDue) {
			s.Curr = qName
			return
		}
	}
	s.Curr = s.init
}
