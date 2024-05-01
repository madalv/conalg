package util

import (
	"context"
	"sync"

	"log/slog"

	"github.com/madalv/conalg/config"
)

type Publisher[T any] interface {
	Subscribe() <-chan T
	CancelSubscription(<-chan T)
	Publish(val T)
}

type broadcastServer[T any] struct {
	listeners []chan T
	mutex     sync.RWMutex
}

func (s *broadcastServer[T]) Subscribe() <-chan T {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newListener := make(chan T, 10000)
	s.listeners = append(s.listeners, newListener)
	slog.Debug("Added listener", config.NR_LIST, len(s.listeners))
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
	slog.Debug("Removed listener", config.NR_LIST, len(s.listeners))
}

func (s *broadcastServer[T]) Publish(val T) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// slog.Info("->>>", val)
	for _, listener := range s.listeners {

		if listener != nil {
			select {
			case listener <- val:
			default:
				slog.Error("Could not send to listener, it is full")
			}
		} else {
			slog.Warn("Channel nil")
		}
	}
}

func NewBroadcastServer[T any](ctx context.Context, source chan T) Publisher[T] {
	service := &broadcastServer[T]{
		listeners: make([]chan T, 0),
		mutex:     sync.RWMutex{},
	}

	return service
}
