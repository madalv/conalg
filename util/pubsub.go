package util

import (
	"context"
	"sync"

	"github.com/gookit/slog"
)

type Publisher[T any] interface {
	Subscribe() <-chan T
	CancelSubscription(<-chan T)
	Publish(val T)
}

type broadcastServer[T any] struct {
	source    chan T
	listeners []chan T
	mutex     sync.Mutex
}

func (s *broadcastServer[T]) Subscribe() <-chan T {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newListener := make(chan T)
	s.listeners = append(s.listeners, newListener)
	return newListener
}

func (s *broadcastServer[T]) CancelSubscription(channel <-chan T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for i, ch := range s.listeners {
		if ch == channel {
			s.listeners[i] = s.listeners[len(s.listeners)-1]
			s.listeners = s.listeners[:len(s.listeners)-1]
			close(ch)
			break
		}
	}
}

func (s *broadcastServer[T]) Publish(val T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, listener := range s.listeners {
		select {
		case listener <- val:
		default:
			slog.Warn("Listener is not ready")
		}
	}
}

func NewBroadcastServer[T any](ctx context.Context, source chan T) Publisher[T] {
	service := &broadcastServer[T]{
		source:    source,
		listeners: make([]chan T, 0),
		mutex:     sync.Mutex{},
	}

	return service
}

