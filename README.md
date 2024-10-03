# spaced

Beat the forgetting curve by reviewing topics using spaced repetition schedule.

There are many apps that help with spaced repetition of things to be memorized
but all of them that I have come across also require that you create these
"items" as index cards in the app.

I wanted a way to add a topic to the app and whenever I run the app, it lists
all the topics which are due for repetition.  It does not care where you refer
to the task.  It may be your text book, or your notebook, or a document
somewhere — it is up to you to manage the references.  The app only tells you
which topic is due for revision.  If I mark the topic performance as good, then
it will reschedule it based on the spaced repetition schedule, if I mark the
performance for that topic as bad, then it resets the schedule and starts from
beginning of the schedule.  So, I made the app which just does that, and
'spaced' is that app.

You can clone the repository and create a binary with `go install` or run the
code directly with `go run`.  When run for the first time, it will ask you for
the location of the data folder, and then it will look for or create new users
in that folder.

The repetition schedule is 0, 1, 3, 7, 21, 30, 45, 60 days from the time of task
creation.  After the schedule is up, the task is forgotten for ever.  You can
still see it in the data file, but the app will not list it in the 'due tasks'
list.

## Config file location and format

- File name: `<user_config_home>/spaced/spacedrc`
- Format:
	path=`<path for data files>`

## Data file location and format

- File name: `<user>.srs`
- Location: <spacedrc.path>
- Format: createdAt|updatedAt|nextInterval|subject|task

# Tasks

## todo

- List all tasks (including expired, active, and due)
- List all tasks for a subject

## done
- Make subject selection for new tasks an option rather than free text

