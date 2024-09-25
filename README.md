# spaced

Beat the forgetting curve by reviewing topics using spaced repetition schedule.

## Design

### Config file

- File name: .config/srs/srsrc
- Format:
	- comments start with #
	- brief description in the comment
	- intervals: 0,1,3,7,21,30,45,60
	- datapath: <path for data files>

### Data file

- File name: <user>.srs
- Format: create_date|last_date|next_interval|subject|task 

### Flow

- Show user names based on the user files
- Allow selection of a user or offer to create new user
- Open or create user file
- START: List active tasks and provide option to select or add task or quit
	- quit: go to END
	- add: take input for new task, go to START
	- select: provide options good, bad, skip, delete selected task
		- good: change last date to today, and set next interval
		- bad: change last date to today, and set first interval
		- skip: do not change anything
		- delete: ask for confirmation
			- yes: soft delete task (next_interval= -1)
			- no: do nothing
- Save file
- Go to START
- END
