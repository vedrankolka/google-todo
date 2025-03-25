package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

const (
	defaultPath         = ".todo/credentials.json"
	defaultTaskListName = "My Tasks"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// TODO: certainly we can offer a way to set cerdentials file path.
		log.Fatalf("unable to get home directory: %v", err)
	}
	keyFilePath := filepath.Join(homeDir, defaultPath)

	b, err := os.ReadFile(keyFilePath)
	if err != nil {
		log.Fatalf("unable to read key file %q: %v", keyFilePath, err)
	}

	scopes := []string{"https://www.googleapis.com/auth/tasks"}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("unable to create JWT configuration: %v", err)
	}

	client := makeOauthClient(config)

	ctx := context.Background()
	srv, err := tasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("unable to create tasks service: %v", err)
	}

	taskLists, err := srv.Tasklists.List().Do()
	if err != nil {
		log.Fatalf("unable to get task lists: %v", err)
	}

	taskListId := ""
	for _, taskList := range taskLists.Items {
		fmt.Printf("Task list: [%s] %s\n", taskList.Id, taskList.Title)
		if taskList.Title == defaultTaskListName {
			taskListId = taskList.Id
		}
	}

	fmt.Println("Task list ID:", taskListId)

	if taskListId == "" {
		log.Fatalf("unable to find task list: %s", defaultTaskListName)
	}

	task := parseTask()

	createdTask, err := srv.Tasks.Insert(taskListId, task).Do()
	if err != nil {
		log.Fatalf("unable to create task: %v", err)
	}

	fmt.Printf("Created task: %v\n", createdTask)
}

func parseTask() *tasks.Task {
	taskTitle := strings.Join(os.Args[1:], " ")

	return &tasks.Task{
		Title:  taskTitle,
		Status: "needsAction",
	}
}
