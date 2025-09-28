package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func IsAllResolved(tasks map[int64]*Task) bool {
	allResolved := true

	for _, task := range tasks {
		if !task.IsResolved() {
			allResolved = false
			break
		}
	}

	return allResolved
}

func SortTasks(messages []*Messge) []*Messge {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})
	return messages
}

func GetUserDataFromCtx(ctx context.Context) (int, string) {
	currentUserId := ctx.Value(UserID("userId")).(int)
	userTag := ctx.Value(UserTag("userTag")).(string)

	return currentUserId, userTag
}

func GetTaskIdFromCommand(command string) (int64, error) {
	patterns := []string{"/assign_", "/unassign_", "/resolve_"}

	var taskID string

	for _, pattern := range patterns {
		if strings.Contains(command, pattern) {
			taskID = strings.TrimSpace(strings.ReplaceAll(command, pattern, ""))
		}
	}

	if len(taskID) == 0 {
		return 0, fmt.Errorf("неизвестная комманда")
	}

	i, err := strconv.ParseInt(taskID, 10, 64)

	if err != nil {
		return i, err
	}

	return i, nil
}
