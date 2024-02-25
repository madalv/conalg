package util

import (
	"context"
)

type Publisher[T any] interface {
	Subscribe() <-chan T
	CancelSubscription(<-chan T)
	Publish(val T)
}

type broadcastServer[T any] struct {
	source         chan T
	listeners      []chan T
	addListener    chan chan T
	removeListener chan (<-chan T)
}

func (s *broadcastServer[T]) Subscribe() <-chan T {
	newListener := make(chan T)
	s.addListener <- newListener
	return newListener
}

func (s *broadcastServer[T]) CancelSubscription(channel <-chan T) {
	s.removeListener <- channel
}

func (s *broadcastServer[T]) Publish(val T) {
	s.source <- val
}

func NewBroadcastServer[T any](ctx context.Context, source chan T) Publisher[T] {
	service := &broadcastServer[T]{
		source:         source,
		listeners:      make([]chan T, 0),
		addListener:    make(chan chan T),
		removeListener: make(chan (<-chan T)),
	}
	go service.serve(ctx)
	return service
}

func (s *broadcastServer[T]) serve(ctx context.Context) {
	defer func() {
		for _, listener := range s.listeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addListener:
			s.listeners = append(s.listeners, newListener)
		case listenerToRemove := <-s.removeListener:
			for i, ch := range s.listeners {
				if ch == listenerToRemove {
					s.listeners[i] = s.listeners[len(s.listeners)-1]
					s.listeners = s.listeners[:len(s.listeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.source:
			if !ok {
				return
			}
			for _, listener := range s.listeners {
				if listener != nil {
					select {
					case listener <- val:
					case <-ctx.Done():
						return
					}

				}
			}
		}
	}
}
