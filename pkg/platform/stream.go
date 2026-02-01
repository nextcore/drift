package platform

import "github.com/go-drift/drift/pkg/errors"

// Stream provides a multi-subscriber broadcast pattern for platform events.
// Unlike raw channels, multiple listeners can receive all events independently.
// Use Listen to subscribe and the returned function to unsubscribe.
type Stream[T any] struct {
	eventChannel *EventChannel
	channelName  string
	parser       func(data any) (T, error)
}

// Listen subscribes to events and returns an unsubscribe function.
// The handler is called for each event. Parse errors are reported via errors.Report.
// Call the returned function to stop receiving events.
func (s *Stream[T]) Listen(handler func(T)) (unsubscribe func()) {
	sub := s.eventChannel.Listen(EventHandler{
		OnEvent: func(data any) {
			val, err := s.parser(data)
			if err != nil {
				errors.Report(&errors.DriftError{
					Op:      "stream.parse",
					Kind:    errors.KindParsing,
					Channel: s.channelName,
					Err:     err,
				})
				return
			}
			handler(val)
		},
		OnError: func(err error) {
			errors.Report(&errors.DriftError{
				Op:      "stream.error",
				Kind:    errors.KindPlatform,
				Channel: s.channelName,
				Err:     err,
			})
		},
	})
	return sub.Cancel
}

// NewStream creates a Stream wrapping an EventChannel.
// The parser converts raw event data to the typed value, returning error on parse failure.
func NewStream[T any](name string, channel *EventChannel, parser func(data any) (T, error)) *Stream[T] {
	return &Stream[T]{
		eventChannel: channel,
		channelName:  name,
		parser:       parser,
	}
}
