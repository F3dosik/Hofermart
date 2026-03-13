package worker

import (
	"container/heap"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobHeap(t *testing.T) {
	now := time.Now()

	j1 := &ScheduleJob{Number: "1", RunAt: now.Add(2 * time.Second)}
	j2 := &ScheduleJob{Number: "2", RunAt: now.Add(1 * time.Second)}
	j3 := &ScheduleJob{Number: "3", RunAt: now.Add(3 * time.Second)}

	h := &JobHeap{}
	heap.Init(h)

	heap.Push(h, j1)
	heap.Push(h, j2)
	heap.Push(h, j3)

	require.Equal(t, 3, h.Len())

	first := heap.Pop(h).(*ScheduleJob)
	second := heap.Pop(h).(*ScheduleJob)
	third := heap.Pop(h).(*ScheduleJob)

	assert.Equal(t, "2", first.Number)
	assert.Equal(t, "1", second.Number)
	assert.Equal(t, "3", third.Number)
}

func TestScheduler_Schedule(t *testing.T) {
	jobChan := make(chan *ScheduleJob, 1)

	s := NewScheduler(jobChan).(*scheduler)

	job := &ScheduleJob{
		Number: "123",
		RunAt:  time.Now(),
	}

	go func() {
		s.Schedule(job)
	}()

	select {
	case j := <-s.sheduleChan:
		assert.Equal(t, "123", j.Number)
	case <-time.After(time.Second):
		t.Fatal("job not scheduled")
	}
}

func TestScheduler_Run(t *testing.T) {
	jobChan := make(chan *ScheduleJob, 1)

	s := NewScheduler(jobChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Run(ctx)

	job := &ScheduleJob{
		Number: "job1",
		RunAt:  time.Now().Add(50 * time.Millisecond),
	}

	s.Schedule(job)

	select {
	case j := <-jobChan:
		assert.Equal(t, "job1", j.Number)
	case <-time.After(1 * time.Second):
		t.Fatal("job was not executed")
	}
}

func TestScheduler_Run_ContextCancel(t *testing.T) {
	jobChan := make(chan *ScheduleJob)

	s := NewScheduler(jobChan)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		s.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop")
	}
}
