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
	CreateTime   time.Time
	UpdateTime   time.Time
	NextInterval int
	Subject      string
	Name         string
}

// New parses a string int the format create_date|last_date|next_interval|subject|task to create a Task
func New(line string) (t *Task, err error) {
	t = &Task{}
	tokens := strings.Split(line, SEP)
	if len(tokens) != 5 {
		err = errors.New("bad data line: " + line)
		return
	}
	if t.CreateTime, err = time.Parse(layout, tokens[0]); err != nil {
		return
	}
	if t.UpdateTime, err = time.Parse(layout, tokens[1]); err != nil {
		return
	}
	if t.NextInterval, err = strconv.Atoi(tokens[2]); err != nil {
		return
	}
	t.Subject = tokens[3]
	t.Name = tokens[4]
	return
}

// New parses a string int the format create_date|last_date|next_interval|subject|task to create a Task
func (t *Task) String() string {
	var s strings.Builder
	s.WriteString(t.CreateTime.Format(layout))
	s.WriteString(SEP)
	s.WriteString(t.UpdateTime.Format(layout))
	s.WriteString(SEP)
	s.WriteString(strconv.Itoa(t.NextInterval))
	s.WriteString(SEP)
	s.WriteString(t.Subject)
	s.WriteString(SEP)
	s.WriteString(t.Name)
	return s.String()
}

func (t *Task) Description() string {
	var s strings.Builder
	s.WriteString("[")
	s.WriteString(t.Subject)
	s.WriteString("] ")
	s.WriteString(t.Name)
	return s.String()
}

var user string
var path string
var layout = time.RFC3339 // "2006-01-02T15:04:05Z07:00"
var alltasks []*Task
var activetasks []*Task
var intervals = []int{0, 1, 3, 7, 21, 30, 45, 60}

const SEP = "|" // field separator in the srs data file
const COM = '#' // a line starting with this character will be ignored for parsing

func main() {
	loadConfig()
	if user = User(); user == "" {
		return
	}
	fmt.Println("User name is:", user)
	filename := path + "/" + user + ".srs"
	if err := Parse(filename); err != nil {
		log.Fatalf("error while parsing data file: %v", err)
	}
	for ShowTasks(activetasks) {
		if err := WriteTasks(filename, alltasks); err != nil {
			log.Fatalf("error while writing data file: %v", err)
		}
		if err := Parse(filename); err != nil {
			log.Fatalf("error while parsing data file: %v", err)
		}
		fmt.Println()
	}
}

// User asks user to select an existing user or create new and returns it
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
			// ignore empty lines and comments
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

func ShowTasks(tasks []*Task) bool {
	if len(tasks) > 0 {
		fmt.Println("tasks due:")
	}
	for i, t := range tasks {
		fmt.Printf("%d. [%s] %s\n", i+1, t.Subject, t.Name)
	}
	for {
		fmt.Print("select task [<sno> | (a)dd new task | (q)uit]: ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "q" {
			return false
		} else if text == "a" {
			t := &Task{
				CreateTime:   time.Now(),
				UpdateTime:   time.Now(),
				NextInterval: 0,
			}
			fmt.Println("add new task: ")
			fmt.Print("enter subject: ")
			text, _ = reader.ReadString('\n')
			t.Subject = strings.TrimSpace(text)
			fmt.Print("enter task name: ")
			text, _ = reader.ReadString('\n')
			t.Name = strings.TrimSpace(text)
			alltasks = append(alltasks, t)
			activetasks = append(activetasks, t)
			return true
		} else {
			if srno, err := strconv.Atoi(text); err == nil {
				if srno > 0 && srno <= len(tasks) {
					current := tasks[srno-1]
					fmt.Println("updating task:", current.Description())
					fmt.Print("how did it go? (g)ood, (b)ad, (d)elete task, (q)uit: ")
					text, _ = reader.ReadString('\n')
					text = strings.TrimSpace(text)
					switch text {
					case "g":
						current.NextInterval = NextInterval(current.NextInterval)
					case "b":
						current.NextInterval = intervals[0]
					case "d":
						fmt.Print("are you sure? (y)es delete, (n)o cancel: ")
						text, _ = reader.ReadString('\n')
						text = strings.TrimSpace(text)
						if text == "y" {
							for i, t := range alltasks {
								if t == current {
									alltasks = append(alltasks[:i], alltasks[i+1:]...)
									break
								}
							}
							for i, t := range activetasks {
								if t == current {
									activetasks = append(activetasks[:i], activetasks[i+1:]...)
									break
								}
							}
							fmt.Println("deleted:", current.Description())
						}
					case "q":
						return false
					default:
						fmt.Println("unknown option:", text)
						continue
					}
					return true
				} else {
					fmt.Println("incorrect choice:", text)
				}
			}
		}
	}
}

func WriteTasks(filename string, tasks []*Task) (err error) {
	// Create a file
	var file *os.File
	if file, err = os.Create(filename); err != nil {
		return
	}
	defer file.Close()

	// Write comment
	if _, err = file.WriteString(fmt.Sprintf("# last updated: %s", time.Now().Format(layout)) + "\n"); err != nil {
		return
	}
	// Write tasks
	for _, t := range alltasks {
		if _, err = file.WriteString(t.String() + "\n"); err != nil {
			return
		}
	}
	return
}

func loadConfig() {
	path = "/home/tapan/tmp"
}

func Parse(fname string) (err error) {
	alltasks = []*Task{}
	activetasks = []*Task{}
	// Open the file
	var file *os.File
	if file, err = os.Open(fname); err != nil {
		return
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read line by line
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		// ignore empty lines and comments
		if len(line) == 0 {
			continue
		} else if line[0] == COM {
			fmt.Println(line)
			continue
		}
		var task *Task
		if task, err = New(line); err != nil {
			return
		}
		alltasks = append(alltasks, task)
		if task.NextInterval >= 0 && task.UpdateTime.AddDate(0, 0, task.NextInterval).Before(time.Now()) {
			activetasks = append(activetasks, task)
		}
	}

	// Check for errors
	err = scanner.Err()
	return
}

func NextInterval(n int) int {
	if n < 0 {
		return n
	}
	for _, num := range intervals {
		if num > n {
			return num
		}
	}
	return -1
}
