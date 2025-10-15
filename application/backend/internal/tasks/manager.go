package tasks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type TaskFunc func(ctx context.Context) error

type Task struct {
	Name       string
	Timeout    time.Duration
	CycleDelay *time.Duration
	Do         TaskFunc
}

type TaskManager struct {
	Tasks        []Task
	taskControls map[string]chan bool
	addTaskChan  chan Task
	stopChan     chan bool
	running      bool
	mu           sync.RWMutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks:        make([]Task, 0),
		taskControls: make(map[string]chan bool),
		addTaskChan:  make(chan Task),
		stopChan:     make(chan bool),
		running:      false,
	}
}

func (tm *TaskManager) Add(task Task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		// If manager is running, send task through channel
		select {
		case tm.addTaskChan <- task:
		default:
			log.Printf("Failed to add task %s - channel full", task.Name)
		}
	} else {
		// If not running, add directly to slice
		tm.Tasks = append(tm.Tasks, task)
	}
}

func (tm *TaskManager) GetTaskByName(name string) *Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	for _, task := range tm.Tasks {
		if task.Name == name {
			return &task
		}
	}
	return nil
}

func (tm *TaskManager) startTask(task Task) {
	stopChan := make(chan bool)
	tm.taskControls[task.Name] = stopChan

	go func(task Task, stop chan bool) {
		for {
			select {
			case <-stop:
				return
			default:
				ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)

				fmt.Println("Starting task:", task.Name)
				err := task.Do(ctx)
				if err != nil {
					fmt.Println("Task", task.Name, "failed:", err)
				} else {
					fmt.Println("Task", task.Name, "completed successfully")
				}

				cancel()

				// Use per-task cycle delay
				if task.CycleDelay != nil {
					time.Sleep(*task.CycleDelay)
				} else {
					stop <- true
				}
			}
		}
	}(task, stopChan)
}

func (tm *TaskManager) Run() {
	tm.mu.Lock()
	tm.running = true
	tm.mu.Unlock()

	log.Println("Running tasks")

	// Start existing tasks
	for _, task := range tm.Tasks {
		tm.startTask(task)
	}

	// Main loop to handle new task additions and stop signals
	for {
		select {
		case newTask := <-tm.addTaskChan:
			tm.mu.Lock()
			tm.Tasks = append(tm.Tasks, newTask)
			tm.mu.Unlock()

			fmt.Printf("Adding new task: %s\n", newTask.Name)
			tm.startTask(newTask)

		case <-tm.stopChan:
			tm.mu.Lock()
			tm.running = false
			tm.mu.Unlock()

			// Stop all running tasks
			for name, stopChan := range tm.taskControls {
				fmt.Printf("Stopping task: %s\n", name)
				stopChan <- true
			}
			return
		}
	}
}

func (tm *TaskManager) Stop() {
	tm.mu.RLock()
	if tm.running {
		tm.stopChan <- true
	}
	tm.mu.RUnlock()
}
