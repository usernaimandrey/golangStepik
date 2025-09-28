package main

import (
	"fmt"
	"sync/atomic"
)

type Storage struct {
	Tasks      map[int64]*Task
	NextTaskId int64
}

func (s *Storage) Inc() {
	atomic.AddInt64(&s.NextTaskId, 1)
}

func NewStorage() *Storage {
	return &Storage{
		Tasks:      map[int64]*Task{},
		NextTaskId: 1,
	}
}

func (s *Storage) InsertTask(t *Task) {
	if _, exists := s.Tasks[t.ID]; exists {
		return
	}

	s.Tasks[t.ID] = t
}

func (s *Storage) TaskFindByID(id int64) (*Task, error) {
	task, ok := s.Tasks[id]

	if !ok {
		return task, fmt.Errorf("task with id: %d is not found", id)
	}

	return task, nil
}

type TaskState string

type Task struct {
	ID       int64
	Title    string
	Reporter *User
	Assegnee *User
	State    TaskState
}

func NewTask(id int64, title string, reporter *User) *Task {
	return &Task{
		ID:       id,
		Title:    title,
		Reporter: reporter,
		State:    "new",
	}
}

func (t *Task) Assign(assegnee *User) (*Task, error) {
	if t.State == "resolved" {
		return t, fmt.Errorf("completed tasks cannot be assigned")
	}

	t.Assegnee = assegnee
	t.State = "appointment"
	return t, nil
}

func (t *Task) Resolving() *Task {
	t.State = "resolved"
	return t
}

func (t *Task) IsAssign() bool {
	return t.State == "appointment"
}

func (t *Task) IsResolved() bool {
	return t.State == "resolved"
}

func (t *Task) Unassign(userId int64) error {
	msg := "Задача не на вас"
	if t.Assegnee.ID != userId {
		return fmt.Errorf(msg)
	}

	t.Assegnee = nil
	t.State = "new"
	return nil
}

type User struct {
	ID  int64
	Tag string
}

func NewUser(id int64, tag string) *User {
	return &User{ID: id, Tag: fmt.Sprintf("@%s", tag)}
}

func (s *Storage) AllTasks() map[int64]*Task {
	return s.Tasks
}
