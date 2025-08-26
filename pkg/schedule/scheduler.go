// pkg/schedule/scheduler.go
package schedule

import (
	"log"
	"time"
)

type Task func()

type Scheduler struct {
	tasks []scheduledTask
}

type scheduledTask struct {
	task      Task
	interval  time.Duration
	lastRun   time.Time
	isRunning bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Add(task Task, interval time.Duration) {
	s.tasks = append(s.tasks, scheduledTask{
		task:     task,
		interval: interval,
		lastRun:  time.Time{}, // Zero time, will run immediately
	})
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.runDueTasks()
	}
}

func (s *Scheduler) runDueTasks() {
	now := time.Now()

	for i := range s.tasks {
		task := &s.tasks[i]

		// Skip if task is already running or not due yet
		if task.isRunning || now.Before(task.lastRun.Add(task.interval)) {
			continue
		}

		// Mark task as running and run it in a goroutine
		task.isRunning = true
		task.lastRun = now

		go func(t *scheduledTask) {
			defer func() {
				t.isRunning = false
				if r := recover(); r != nil {
					log.Printf("Task panicked: %v", r)
				}
			}()

			log.Printf("Running scheduled task")
			t.task()
		}(task)
	}
}

// Convenience methods for common intervals
func (s *Scheduler) EveryMinute(task Task) {
	s.Add(task, time.Minute)
}

func (s *Scheduler) EveryFiveMinutes(task Task) {
	s.Add(task, 5*time.Minute)
}

func (s *Scheduler) EveryHour(task Task) {
	s.Add(task, time.Hour)
}

func (s *Scheduler) Daily(task Task) {
	s.Add(task, 24*time.Hour)
}
