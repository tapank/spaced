package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	NextInterval time.Duration
	Subject      string
	Name         string
}

// New parses string of format 'create_date|last_date|next_interval|subject|task' to create a Task
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
	if nextInterval, e := strconv.Atoi(tokens[2]); e != nil {
		if t.NextInterval, err = time.ParseDuration(tokens[2]); err != nil {
			return
		}
	} else {
		t.NextInterval = time.Duration(nextInterval) * DAY
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
	s.WriteString(FormatDuration(t.NextInterval))
	s.WriteString(SEP)
	s.WriteString(t.Subject)
	s.WriteString(SEP)
	s.WriteString(t.Name)
	return s.String()
}

func (t *Task) Description() string {
	due := time.Since(t.UpdateTime.Add(t.NextInterval))
	var s strings.Builder
	s.WriteString("[")
	s.WriteString(fmt.Sprintf("%2dd %02dh", Days(due), Hours(due)))
	s.WriteString("]")
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
var duetasks []*Task
var upcomingtasks []*Task
var subjects []string

const DAY = time.Hour * 24

// intervals must have at least one element or the new task creation will fail at runtime
var intervals = []time.Duration{12 * time.Hour, 1 * DAY, 3 * DAY, 7 * DAY, 21 * DAY, 30 * DAY, 45 * DAY, 60 * DAY}

const SEP = "|" // field separator in the srs data file
const COM = '#' // a line starting with this character will be ignored for parsing
const configdir = "spaced"
const configfile = "spacedrc"

func main() {
	loadConfig()
	if user = User(); user == "" {
		return
	}
	filename := path + "/" + user + ".srs"
	if err := Parse(filename); err != nil {
		log.Fatalf("error parsing data file: %v", err)
	}
	for ShowTasks() {
		if err := WriteTasks(filename, alltasks); err != nil {
			log.Fatalf("error writing data file: %v", err)
		}
		fmt.Println()
		if err := Parse(filename); err != nil {
			log.Fatalf("error parsing data file: %v", err)
		}
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
		log.Fatalf("error finding data files: %v", err)
	}

	users := []string{}

	// extract usernames from file paths
	for _, path := range userFiles {
		file := filepath.Base(path)
		file, _ = strings.CutSuffix(file, ".srs")
		users = append(users, file)
	}

	clearscreen()
	if len(users) == 0 {
		fmt.Println("no user files found")
		fmt.Print("\nenter your choice [(a)dd new user | (q)uit]: ")
	} else {
		fmt.Println("select user:")
		for i, username := range users {
			fmt.Println(i+1, username)
		}
		fmt.Print("\nenter your choice [<sno> | (a)dd new user | (q)uit]: ")
	}

	switch text := GetInput(""); text {
	case "q":
		return ""
	case "a":
		text = GetInput("enter user name: ")
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
	default:
		if index, err := strconv.Atoi(text); err != nil {
			log.Fatalf("invalid choice: %s", text)
		} else {
			index -= 1
			if index >= 0 && index < len(users) {
				return users[index]
			}
		}
	}
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
			return "", errors.New("only alphabets, digits, hypyens, underscores allowed in user name")
		}
	}
	return s, nil
}

func ShowTasks() bool {
	clearscreen()
	fmt.Printf("tasks for '%s' (total: %d, due: %d):\n", user, len(alltasks), len(duetasks))
	fmt.Println("\ndue:")
	if len(duetasks) == 0 {
		fmt.Println("-")
	}
	for i, t := range duetasks {
		fmt.Printf("%d. %s\n", i+1, t.Description())
	}

	fmt.Println("\ncoming up:")
	if len(upcomingtasks) > 0 {
	} else {
		fmt.Println("-")
	}
	for _, t := range upcomingtasks {
		fmt.Printf(".. %s\n", t.Description())
	}

	for {
		switch text := GetInput("\nselect task [<sno> | (a)dd new task | (q)uit]: "); text {
		case "q":
			return false
		case "a":
			t := &Task{
				CreateTime:   time.Now(),
				UpdateTime:   time.Now(),
				NextInterval: intervals[0],
			}
			fmt.Println("add new task: ")
			for {
				subjectChoice := GetInput(subjectsList())
				if subjectChoice == "a" {
					t.Subject = strings.ToLower(GetInput("enter new subject: "))
					break
				} else if subjectIndex, err := strconv.Atoi(subjectChoice); err != nil {
					fmt.Println("invalid choice:", subjectChoice)
				} else if subjectIndex < 1 || subjectIndex > len(subjects) {
					fmt.Println("invalid choice:", subjectChoice)
				} else {
					t.Subject = subjects[subjectIndex-1]
					break
				}
			}
			t.Name = GetInput("enter task name: ")
			alltasks = append(alltasks, t)
			duetasks = append(duetasks, t)
			return true
		default:
			if srno, err := strconv.Atoi(text); err == nil {
				if srno > 0 && srno <= len(duetasks) {
					current := duetasks[srno-1]
					fmt.Println("updating task:", current.Description())
					switch GetInput("how did it go? (g)ood, (b)ad, (d)elete task, (q)uit: ") {
					case "g":
						current.NextInterval = NextInterval(current.NextInterval)
						current.UpdateTime = time.Now()
					case "b":
						current.NextInterval = intervals[0]
						current.UpdateTime = time.Now()
					case "d":
						if text = GetInput("are you sure? (y)es delete, (n)o cancel: "); text == "y" {
							for i, t := range alltasks {
								if t == current {
									alltasks = append(alltasks[:i], alltasks[i+1:]...)
									break
								}
							}
							for i, t := range duetasks {
								if t == current {
									duetasks = append(duetasks[:i], duetasks[i+1:]...)
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
	// Create or erase data file
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
	confighome, _ := os.UserConfigDir()
	if _, err := os.Stat(confighome + "/" + configdir + "/" + configfile); err != nil {
		fmt.Println("config file does not exist")
		createConfig()
	}

	// Open config file
	var file *os.File
	var err error
	if file, err = os.Open(confighome + "/" + configdir + "/" + configfile); err != nil {
		log.Fatalf("error opening config file: %v", err)
	}
	defer file.Close()

	var scanner *bufio.Scanner
	for scanner = bufio.NewScanner(file); scanner.Scan(); {
		// ignore empty lines and comments
		if line := strings.TrimSpace(scanner.Text()); len(line) == 0 {
			continue
		} else if line[0] == COM {
			fmt.Println(line)
			continue
		} else {
			if strings.HasPrefix(line, "path=") {
				path = strings.TrimPrefix(line, "path=")
			}
		}
	}

	// Check for errors
	if err = scanner.Err(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}
	if path == "" {
		log.Fatalf("path not found in config file")
	}
}

func createConfig() {
	var err error
	home, _ := os.UserConfigDir()
	err = os.Mkdir(home+"/"+configdir, 0755)
	if err != nil {
		log.Fatalf("error creating config dir: %v", err)
	}

	var file *os.File
	if file, err = os.Create(home + "/" + configdir + "/" + configfile); err != nil {
		log.Fatalf("error creating config file: %v", err)
	}
	defer file.Close()

	text := GetInput("enter path to data folder: ")
	// Write config
	if _, err = file.WriteString("path=" + text + "\n"); err != nil {
		log.Fatalf("error writing config file: %v", err)
	}
}

func Parse(fname string) (err error) {
	alltasks = []*Task{}
	duetasks = []*Task{}
	upcomingtasks = []*Task{}
	subjects = []string{}

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
		line := strings.TrimSpace(scanner.Text())
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
		if !slices.Contains(subjects, task.Subject) {
			subjects = append(subjects, task.Subject)
		}
		if task.NextInterval >= 0 && task.UpdateTime.Add(task.NextInterval).Before(time.Now()) {
			duetasks = append(duetasks, task)
		} else if task.NextInterval >= 0 && task.UpdateTime.Add(task.NextInterval).Before(time.Now().Add(2*DAY)) {
			upcomingtasks = append(upcomingtasks, task)
		}
	}
	return scanner.Err()
}

func NextInterval(duration time.Duration) time.Duration {
	if duration < 0 {
		return duration
	}
	for _, d := range intervals {
		if d > duration {
			return d
		}
	}
	return -1
}

func GetInput(msg string) string {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func clearscreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func subjectsList() string {
	var sb strings.Builder
	sb.WriteString("select subject")
	for i, s := range subjects {
		sb.WriteString(" (" + strconv.Itoa(i+1) + ") ")
		sb.WriteString(s)
	}
	sb.WriteString(" (a)dd new subject: ")
	return sb.String()
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "-1"
	} else if d >= DAY || d < time.Hour {
		return strconv.Itoa(int(d / DAY))
	}
	return strconv.Itoa(int(d/time.Hour)) + "h"
}

func Days(d time.Duration) int {
	if d < 0 {
		d = -d
	}
	return int(d / DAY)
}

func Hours(d time.Duration) int {
	if d < 0 {
		d = -d
	}
	return int((d % DAY) / time.Hour)
}
