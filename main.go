package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

const (
	clientID = "586546061595-9p09kjl3f97936dfo1b42tv3itm0o8sd.apps.googleusercontent.com"
	// clientSecret is not a secret for installed apps, as stated in Google's docs:
	// https://developers.google.com/identity/protocols/oauth2#installed
	clientSecret        = "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"
	defaultTaskListName = "My Tasks"
	redirectURL         = "http://localhost:8090"
	scope               = "https://www.googleapis.com/auth/tasks"
)

func main() {
	listName := flag.String("list", defaultTaskListName, "task list name")

	flag.Parse()

	task, err := parseTask(flag.Args())
	if err != nil {
		log.Fatalf("unable to parse task: %v", err)
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{scope},
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to get user home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "todo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("unable to create config directory: %v", err)
	}

	tokenPath := filepath.Join(configDir, "token.json")

	client := makeOauthClient(config, tokenPath)

	ctx := context.Background()
	srv, err := tasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("unable to create tasks service: %v", err)
	}

	taskLists, err := srv.Tasklists.List().Do()
	if err != nil {
		log.Fatalf("unable to get task lists: %v", err)
	}

	taskListId, err := findTaskList(taskLists.Items, *listName)
	if err != nil {
		log.Fatalf("unable to find task list: %s", *listName)
	}

	if _, err := srv.Tasks.Insert(taskListId, task).Do(); err != nil {
		log.Fatalf("unable to create task: %v", err)
	}
}

func parseTask(args []string) (*tasks.Task, error) {
	taskTitle := strings.Join(args, " ")

	if taskTitle == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	return &tasks.Task{
		Title:  taskTitle,
		Status: "needsAction",
	}, nil
}

func findTaskList(taskLists []*tasks.TaskList, listName string) (string, error) {
	for _, taskList := range taskLists {
		if taskList.Title == listName {
			return taskList.Id, nil
		}
	}

	return "", fmt.Errorf("unable to find task list: %s", listName)
}
