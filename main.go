package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Task is the item that is repeated as per schedule
type Task struct {
	ID           int
	CreateTime   time.Time
	UpdateTime   time.Time
	NextInterval time.Duration
	Subject      string
	Name         string
}

var tasks map[int]*Task
var user string
var path string

// var layout = time.RFC3339 // "2006-01-02T15:04:05Z07:00"

func User() string {
	// collect file names with .srs extension
	userFiles, err := func(root, pattern string) ([]string, error) {
		var matches []string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
				return err
			} else if matched {
				matches = append(matches, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return matches, nil
	}(path, "*.srs")
	if err != nil {
		log.Fatalf("Encountered error while looking for data files: %v", err)
	}

	users := []string{}

	// extract usernames from file paths
	for _, path := range userFiles {
		file := filepath.Base(path)
		file, _ = strings.CutSuffix(file, ".srs")
		users = append(users, file)
	}

	if len(users) == 0 {
		fmt.Println("no user files found")
		fmt.Print("enter your choice [(a)dd new user | (q)uit]: ")
	} else {
		fmt.Println("Select user:")
		for i, username := range users {
			fmt.Println(i+1, username)
		}
		fmt.Print("enter your choice [<sno> | (a)dd new user | (q)uit]: ")
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "q" {
		return ""
	} else if text == "a" {
		fmt.Print("enter user name: ")
		text, _ = reader.ReadString('\n')
		if name, err := validateUserName(text); err != nil {
			log.Fatalf("invalid use name: %v", err)
		} else if slices.Contains(users, name) {
			log.Fatalf("user already exists: %s", name)
		} else {
			// Create an empty file
			if f, err := os.Create(path + "/" + name + ".srs"); err != nil {
				log.Fatalf("error creating data file: %v", err)
			} else {
				f.Close()
			}
			return name
		}
	}
	if index, err := strconv.Atoi(text); err != nil {
		log.Fatalf("invalid choice: %s", text)
	} else {
		index -= 1
		if index >= 0 && index < len(users) {
			return users[index]
		}
	}
	log.Fatalf("invalid choice: %s", text)
	return ""
}

func validateUserName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return "", errors.New("name cannot be blank")
	}
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
		case (r >= '0' && r <= '9'):
		case r == '_' || r == '-':
		default:
			return "", errors.New("only alphabets, digits, hypyen, underscore allowed")
		}
	}
	return s, nil
}

func Tasks(user string) map[int]*Task {
	return nil
}

func ShowTasks(tasks map[int]*Task) bool {
	return false
}

func WriteTasks(tasks map[int]*Task) {
}

func loadConfig() {
	path = "/home/tapan/tmp"
}

func main() {
	loadConfig()
	if user = User(); user == "" {
		return
	}
	fmt.Println("User name is:", user)
	tasks = Tasks(user)
	for ShowTasks(tasks) {
		WriteTasks(tasks)
	}
}
