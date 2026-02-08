package platform

import (
	"context"
	"time"

	"github.com/go-drift/drift/pkg/errors"
)

// TaskType defines the type of background task.
type TaskType string

const (
	TaskTypeOneTime  TaskType = "one_time"
	TaskTypePeriodic TaskType = "periodic"
	TaskTypeFetch    TaskType = "fetch"
)

// TaskConstraints defines constraints for when a task should run.
type TaskConstraints struct {
	RequiresNetwork          bool
	RequiresUnmeteredNetwork bool
	RequiresCharging         bool
	RequiresIdle             bool
	RequiresStorageNotLow    bool
	RequiresBatteryNotLow    bool
}

// TaskRequest describes a background task to schedule.
type TaskRequest struct {
	ID             string
	TaskType       TaskType
	Tag            string
	Constraints    TaskConstraints
	InitialDelay   time.Duration
	RepeatInterval time.Duration
	Data           map[string]any
}

// BackgroundEvent represents a background task event.
type BackgroundEvent struct {
	TaskID    string
	EventType string
	Data      map[string]any
	Timestamp time.Time
}

// BackgroundService provides background task scheduling and event access.
type BackgroundService struct {
	state  *backgroundServiceState
	events *Stream[BackgroundEvent]
}

// Background is the singleton background service.
var Background *BackgroundService

func init() {
	state := newBackgroundService()
	Background = &BackgroundService{
		state:  state,
		events: NewStream("drift/background/events", state.events, parseBackgroundEventWithError),
	}
}

type backgroundServiceState struct {
	channel *MethodChannel
	events  *EventChannel
}

func newBackgroundService() *backgroundServiceState {
	return &backgroundServiceState{
		channel: NewMethodChannel("drift/background"),
		events:  NewEventChannel("drift/background/events"),
	}
}

// Schedule schedules a background task.
func (b *BackgroundService) Schedule(ctx context.Context, request TaskRequest) error {
	taskType := string(request.TaskType)
	if taskType == "" {
		taskType = string(TaskTypeOneTime)
	}

	_, err := b.state.channel.Invoke("scheduleTask", map[string]any{
		"id":               request.ID,
		"taskType":         taskType,
		"tag":              request.Tag,
		"initialDelayMs":   request.InitialDelay.Milliseconds(),
		"repeatIntervalMs": request.RepeatInterval.Milliseconds(),
		"data":             request.Data,
		"constraints": map[string]any{
			"requiresNetwork":          request.Constraints.RequiresNetwork,
			"requiresUnmeteredNetwork": request.Constraints.RequiresUnmeteredNetwork,
			"requiresCharging":         request.Constraints.RequiresCharging,
			"requiresIdle":             request.Constraints.RequiresIdle,
			"requiresStorageNotLow":    request.Constraints.RequiresStorageNotLow,
			"requiresBatteryNotLow":    request.Constraints.RequiresBatteryNotLow,
		},
	})
	return err
}

// Cancel cancels a scheduled background task.
func (b *BackgroundService) Cancel(ctx context.Context, id string) error {
	_, err := b.state.channel.Invoke("cancelTask", map[string]any{
		"id": id,
	})
	return err
}

// CancelAll cancels all scheduled background tasks.
func (b *BackgroundService) CancelAll(ctx context.Context) error {
	_, err := b.state.channel.Invoke("cancelAllTasks", nil)
	return err
}

// CancelByTag cancels all tasks with the given tag.
func (b *BackgroundService) CancelByTag(ctx context.Context, tag string) error {
	_, err := b.state.channel.Invoke("cancelTasksByTag", map[string]any{
		"tag": tag,
	})
	return err
}

// Complete signals completion of a background task.
func (b *BackgroundService) Complete(ctx context.Context, id string, success bool) error {
	_, err := b.state.channel.Invoke("completeTask", map[string]any{
		"id":      id,
		"success": success,
	})
	return err
}

// IsRefreshAvailable checks if background refresh is available.
func (b *BackgroundService) IsRefreshAvailable(ctx context.Context) (bool, error) {
	result, err := b.state.channel.Invoke("isBackgroundRefreshAvailable", nil)
	if err != nil {
		return false, err
	}
	if m, ok := result.(map[string]any); ok {
		return parseBool(m["available"]), nil
	}
	return false, nil
}

// Events returns a stream of background task events.
func (b *BackgroundService) Events() *Stream[BackgroundEvent] {
	return b.events
}

func parseBackgroundEventWithError(data any) (BackgroundEvent, error) {
	m, ok := data.(map[string]any)
	if !ok {
		return BackgroundEvent{}, &errors.ParseError{
			Channel:  "drift/background/events",
			DataType: "BackgroundEvent",
			Got:      data,
		}
	}
	return BackgroundEvent{
		TaskID:    parseString(m["taskId"]),
		EventType: parseString(m["eventType"]),
		Data:      parseMap(m["data"]),
		Timestamp: parseTime(m["timestamp"]),
	}, nil
}
