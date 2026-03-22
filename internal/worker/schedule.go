package worker

import (
	"container/heap"
	"context"
	"time"
)

type Scheduler interface {
	Schedule(job *ScheduleJob)
	Run(ctx context.Context)
}

type scheduler struct {
	sheduleChan chan *ScheduleJob
	jobChan     chan *ScheduleJob
}

func NewScheduler(jobChan chan *ScheduleJob) Scheduler {
	return &scheduler{sheduleChan: make(chan *ScheduleJob), jobChan: jobChan}
}

func (s *scheduler) Run(ctx context.Context) {
	h := &JobHeap{}
	heap.Init(h)

	var timer *time.Timer

	for {
		var nextTimer <-chan time.Time
		if h.Len() > 0 {
			nextRun := (*h)[0].RunAt
			delay := max(time.Until(nextRun), 0)
			if timer == nil {
				timer = time.NewTimer(delay)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(delay)
			}
			nextTimer = timer.C
		}

		select {
		case job := <-s.sheduleChan:
			heap.Push(h, job)
		case <-nextTimer:
			job := heap.Pop(h).(*ScheduleJob)
			s.jobChan <- job
		case <-ctx.Done():
			return
		}

	}
}

func (s *scheduler) Schedule(job *ScheduleJob) {
	s.sheduleChan <- job
}

type ScheduleJob struct {
	Number  string
	RunAt   time.Time
	Attempt int
	index   int
}

type JobHeap []*ScheduleJob

func (h JobHeap) Len() int { return len(h) }

func (h JobHeap) Less(i, j int) bool {
	return h[i].RunAt.Before(h[j].RunAt)
}

func (h JobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *JobHeap) Push(x any) {
	n := len(*h)
	job := x.(*ScheduleJob)
	job.index = n
	*h = append(*h, job)
}

func (h *JobHeap) Pop() any {
	old := *h
	n := len(old)
	job := old[n-1]
	job.index = -1
	*h = old[:n-1]
	return job
}
