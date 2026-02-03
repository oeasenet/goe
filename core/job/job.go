package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.oease.dev/goe/v2/contract"
)

// jobImpl implements the contract.Job interface
type jobImpl struct {
	id          string
	name        string
	queue       string
	payload     any
	status      contract.JobStatus
	attempts    int
	maxAttempts int
	createdAt   time.Time
	scheduledAt time.Time
	startedAt   time.Time
	completedAt time.Time
	lastError   string
	ctx         context.Context
	tags        map[string]string
	timeout     time.Duration
	uniqueKey   string
}

func (j *jobImpl) ID() string                 { return j.id }
func (j *jobImpl) Name() string               { return j.name }
func (j *jobImpl) Queue() string              { return j.queue }
func (j *jobImpl) Payload() any               { return j.payload }
func (j *jobImpl) Status() contract.JobStatus { return j.status }
func (j *jobImpl) Attempts() int              { return j.attempts }
func (j *jobImpl) MaxAttempts() int           { return j.maxAttempts }
func (j *jobImpl) CreatedAt() time.Time       { return j.createdAt }
func (j *jobImpl) ScheduledAt() time.Time     { return j.scheduledAt }
func (j *jobImpl) StartedAt() time.Time       { return j.startedAt }
func (j *jobImpl) CompletedAt() time.Time     { return j.completedAt }
func (j *jobImpl) LastError() string          { return j.lastError }
func (j *jobImpl) Context() context.Context   { return j.ctx }
func (j *jobImpl) Tags() map[string]string    { return j.tags }
func (j *jobImpl) Timeout() time.Duration     { return j.timeout }
func (j *jobImpl) UniqueKey() string          { return j.uniqueKey }

// jobData is the serializable form of a job for storage
type jobData struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Queue       string             `json:"queue"`
	Payload     json.RawMessage    `json:"payload"`
	Status      contract.JobStatus `json:"status"`
	Attempts    int                `json:"attempts"`
	MaxAttempts int                `json:"max_attempts"`
	CreatedAt   int64              `json:"created_at"`
	ScheduledAt int64              `json:"scheduled_at"`
	StartedAt   int64              `json:"started_at,omitempty"`
	CompletedAt int64              `json:"completed_at,omitempty"`
	LastError   string             `json:"last_error,omitempty"`
	Tags        map[string]string  `json:"tags,omitempty"`
	Timeout     int64              `json:"timeout"`
	UniqueKey   string             `json:"unique_key,omitempty"`
}

// newJob creates a new job from a definition
func newJob(def *contract.JobDefinition) (*jobImpl, error) {
	id := uuid.New().String()

	now := time.Now()
	scheduledAt := now
	status := contract.JobStatusPending

	if !def.ScheduledAt.IsZero() {
		scheduledAt = def.ScheduledAt
		status = contract.JobStatusScheduled
	} else if def.Delay > 0 {
		scheduledAt = now.Add(def.Delay)
		status = contract.JobStatusScheduled
	}

	queue := def.Queue
	if queue == "" {
		queue = "default"
	}

	maxAttempts := def.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	timeout := def.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}

	return &jobImpl{
		id:          id,
		name:        def.Name,
		queue:       queue,
		payload:     def.Payload,
		status:      status,
		attempts:    0,
		maxAttempts: maxAttempts,
		createdAt:   now,
		scheduledAt: scheduledAt,
		tags:        def.Tags,
		timeout:     timeout,
		uniqueKey:   def.UniqueKey,
		ctx:         context.Background(),
	}, nil
}

// toJobData converts a jobImpl to jobData for storage
func (j *jobImpl) toJobData() (*jobData, error) {
	payloadBytes, err := json.Marshal(j.payload)
	if err != nil {
		return nil, err
	}

	return &jobData{
		ID:          j.id,
		Name:        j.name,
		Queue:       j.queue,
		Payload:     payloadBytes,
		Status:      j.status,
		Attempts:    j.attempts,
		MaxAttempts: j.maxAttempts,
		CreatedAt:   j.createdAt.UnixMilli(),
		ScheduledAt: j.scheduledAt.UnixMilli(),
		StartedAt:   j.startedAt.UnixMilli(),
		CompletedAt: j.completedAt.UnixMilli(),
		LastError:   j.lastError,
		Tags:        j.tags,
		Timeout:     j.timeout.Milliseconds(),
		UniqueKey:   j.uniqueKey,
	}, nil
}

// fromJobData creates a jobImpl from jobData
func fromJobData(data *jobData) *jobImpl {
	var payload any
	if len(data.Payload) > 0 {
		_ = json.Unmarshal(data.Payload, &payload)
	}

	return &jobImpl{
		id:          data.ID,
		name:        data.Name,
		queue:       data.Queue,
		payload:     payload,
		status:      data.Status,
		attempts:    data.Attempts,
		maxAttempts: data.MaxAttempts,
		createdAt:   time.UnixMilli(data.CreatedAt),
		scheduledAt: time.UnixMilli(data.ScheduledAt),
		startedAt:   time.UnixMilli(data.StartedAt),
		completedAt: time.UnixMilli(data.CompletedAt),
		lastError:   data.LastError,
		tags:        data.Tags,
		timeout:     time.Duration(data.Timeout) * time.Millisecond,
		uniqueKey:   data.UniqueKey,
		ctx:         context.Background(),
	}
}

// marshal converts jobData to JSON bytes
func (d *jobData) marshal() ([]byte, error) {
	return json.Marshal(d)
}

// unmarshalJobData parses JSON bytes into jobData
func unmarshalJobData(data []byte) (*jobData, error) {
	var jd jobData
	if err := json.Unmarshal(data, &jd); err != nil {
		return nil, err
	}
	return &jd, nil
}
