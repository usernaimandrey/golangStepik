package main

import (
	"context"
	"fmt"
	"strings"
)

type Router struct {
	Routes map[string]func(ctx context.Context, db *Storage, msg string) []*Messge
}

type Messge struct {
	Msg string
	ID  int64
}

func NewMessage(msg string, id int64) *Messge {
	return &Messge{Msg: msg, ID: id}
}

func NewRouter() *Router {
	return &Router{
		Routes: map[string]func(ctx context.Context, db *Storage, msg string) []*Messge{
			"/tasks":    Tasks,
			"/new":      NewTasks,
			"/assign":   Assign,
			"/unassign": Unassign,
			"/resolve":  Resolve,
			"/my":       MyTasks,
			"/owner":    OwnerTasks,
		},
	}
}

func (r *Router) SelectRoute(ctx context.Context, db *Storage, msg string) ([]*Messge, error) {
	commands := []string{
		"/tasks", "/new", "/assign", "/unassign", "/resolve", "/my", "/owner",
	}

	var route string
	isFind := false

	for _, command := range commands {
		index := strings.Index(msg, command)
		if index != -1 {
			route = command
			isFind = true
			break
		}
	}

	if !isFind {
		return []*Messge{}, fmt.Errorf("unknown command")
	}

	return r.Routes[route](ctx, db, msg), nil
}

func Tasks(ctx context.Context, db *Storage, msg string) []*Messge {
	tasks := db.AllTasks()

	currentUserId, _ := GetUserDataFromCtx(ctx)

	msgs := []*Messge{}

	if len(tasks) == 0 || IsAllResolved(tasks) {
		msg := NewMessage("Нет задач", int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	delimetr := "\n\t\t\t\t"
	assignDelimetr := "\n\t\t"
	assignTmpl := "%d. %s by %s%sassignee: %s%s/unassign_%d /resolve_%d"
	assignTmplWithoutCommand := "%d. %s by %s%sassignee: %s"
	tmpl := "%d. %s by %s%s/assign_%d"
	responseMessage := ""

	for _, task := range tasks {
		if task.IsResolved() {
			continue
		}

		if task.IsAssign() && currentUserId == int(task.Assegnee.ID) {
			responseMessage += fmt.Sprintf(assignTmpl, task.ID, task.Title, task.Reporter.Tag, assignDelimetr, "я", assignDelimetr, task.ID, task.ID) + "\n\n\t\t"
		} else if task.IsAssign() && currentUserId != int(task.Assegnee.ID) {
			responseMessage += fmt.Sprintf(assignTmplWithoutCommand, task.ID, task.Title, task.Reporter.Tag, assignDelimetr, task.Assegnee.Tag) + "\n\n\t\t"
		} else {
			responseMessage += (fmt.Sprintf(tmpl, task.ID, task.Title, task.Reporter.Tag, delimetr, task.ID) + "\n\n\t\t")
		}

		reporterMessage := NewMessage(strings.TrimSuffix(responseMessage, "\n\n\t\t"), int64(currentUserId))
		msgs = append(msgs, reporterMessage)
	}

	sortedMsgs := SortTasks(msgs)
	return sortedMsgs
}

func NewTasks(ctx context.Context, db *Storage, command string) []*Messge {
	currentUserId, userTag := GetUserDataFromCtx(ctx)

	title := strings.TrimSpace(strings.ReplaceAll(command, "/new", ""))

	reporter := NewUser(int64(currentUserId), userTag)
	newTask := NewTask(db.NextTaskId, title, reporter)

	db.Inc()

	db.InsertTask(newTask)

	msgs := []*Messge{}
	msg := NewMessage(fmt.Sprintf(`Задача "%s" создана, id=%d`, title, newTask.ID), int64(currentUserId))

	msgs = append(msgs, msg)
	return msgs
}

func Assign(ctx context.Context, db *Storage, command string) []*Messge {
	currentUserId, userTag := GetUserDataFromCtx(ctx)

	taskID, err := GetTaskIdFromCommand(command)

	msgs := []*Messge{}

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	task, err := db.TaskFindByID(taskID)

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	newAppointed := NewUser(int64(currentUserId), userTag)
	oldAppointed := task.Assegnee

	tmpl := `Задача "%s" назначена на %s`

	var reporterMsg *Messge

	appointerMsg := NewMessage(fmt.Sprintf(tmpl, task.Title, "вас"), newAppointed.ID)

	if task.IsAssign() {
		reporterMsg = NewMessage(fmt.Sprintf(tmpl, task.Title, newAppointed.Tag), oldAppointed.ID)
	} else {
		reporterMsg = NewMessage(fmt.Sprintf(tmpl, task.Title, newAppointed.Tag), task.Reporter.ID)
	}

	_, err = task.Assign(newAppointed)

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	if task.Reporter.ID == int64(currentUserId) {
		msgs = append(msgs, appointerMsg)
		return msgs
	}

	msgs = append(msgs, []*Messge{appointerMsg, reporterMsg}...)
	return msgs
}

func Unassign(ctx context.Context, db *Storage, command string) []*Messge {
	currentUserId, _ := GetUserDataFromCtx(ctx)

	taskID, err := GetTaskIdFromCommand(command)

	msgs := []*Messge{}

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	task, err := db.TaskFindByID(taskID)

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	err = task.Unassign(int64(currentUserId))

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	appointerMsg := NewMessage(`Принято`, int64(currentUserId))
	reporterMsg := NewMessage(`Задача "написать бота" осталась без исполнителя`, task.Reporter.ID)

	msgs = append(msgs, []*Messge{appointerMsg, reporterMsg}...)

	return msgs
}

func Resolve(ctx context.Context, db *Storage, command string) []*Messge {
	currentUserId, userTag := GetUserDataFromCtx(ctx)

	taskID, err := GetTaskIdFromCommand(command)

	msgs := []*Messge{}

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	task, err := db.TaskFindByID(taskID)

	if err != nil {
		msg := NewMessage(err.Error(), int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	task.Resolving()

	appointerMsg := NewMessage(`Задача "написать бота" выполнена`, int64(currentUserId))
	reporterMsg := NewMessage(fmt.Sprintf(`Задача "написать бота" выполнена @%s`, userTag), task.Reporter.ID)

	msgs = append(msgs, []*Messge{appointerMsg, reporterMsg}...)

	return msgs
}

func MyTasks(ctx context.Context, db *Storage, command string) []*Messge {
	tasks := db.AllTasks()

	currentUserId, _ := GetUserDataFromCtx(ctx)

	msgs := []*Messge{}

	if len(tasks) == 0 || IsAllResolved(tasks) {
		msg := NewMessage("Нет задач", int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	assignTmpl := "%d. %s by %s%s/unassign_%d /resolve_%d"
	responseMessage := ""
	assignDelimetr := "\n\t\t"

	for _, task := range tasks {
		if !task.IsAssign() {
			continue
		}

		if task.Assegnee.ID != int64(currentUserId) {
			continue
		}

		responseMessage += fmt.Sprintf(assignTmpl, task.ID, task.Title, task.Reporter.Tag, assignDelimetr, task.ID, task.ID) + "\n\n\t\t"

		reporterMessage := NewMessage(strings.TrimSuffix(responseMessage, "\n\n\t\t"), int64(currentUserId))
		msgs = append(msgs, reporterMessage)
	}

	sortedMsgs := SortTasks(msgs)
	return sortedMsgs
}

func OwnerTasks(ctx context.Context, db *Storage, command string) []*Messge {
	tasks := db.AllTasks()

	currentUserId, _ := GetUserDataFromCtx(ctx)

	msgs := []*Messge{}

	if len(tasks) == 0 || IsAllResolved(tasks) {
		msg := NewMessage("Нет задач", int64(currentUserId))
		msgs = append(msgs, msg)
		return msgs
	}

	delimetr := "\n\t\t"
	tmpl := "%d. %s by %s%s/assign_%d"
	responseMessage := ""

	for _, task := range tasks {
		if task.Reporter.ID != int64(currentUserId) || task.IsResolved() {
			continue
		}

		responseMessage += (fmt.Sprintf(tmpl, task.ID, task.Title, task.Reporter.Tag, delimetr, task.ID) + "\n\n\t\t")

		reporterMessage := NewMessage(strings.TrimSuffix(responseMessage, "\n\n\t\t"), int64(currentUserId))
		msgs = append(msgs, reporterMessage)
	}

	sortedMsgs := SortTasks(msgs)
	return sortedMsgs
}
