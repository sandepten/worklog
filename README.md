# Work Obsidian Noter

A CLI tool for managing daily work notes in Obsidian. Track your pending and completed work items, review yesterday's tasks, and get AI-powered summaries of your accomplishments.

## Features

- Create daily work notes in Obsidian-compatible markdown format
- Interactive review of pending items from previous days
- AI-powered work summaries using OpenCode server
- Carry forward incomplete tasks to the next day
- Track completed work with checkboxes

## Installation

```bash
# Clone the repository
git clone https://github.com/sandepten/work-obsidian-noter.git
cd work-obsidian-noter

# Install dependencies
go mod tidy

# Build and install (creates symlink in ~/.local/bin)
make install

# Or just build without installing
make build
```

> **Note:** Make sure `~/.local/bin` is in your PATH. Add this to your `~/.bashrc` or `~/.zshrc`:
> ```bash
> export PATH="$HOME/.local/bin:$PATH"
> ```

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the `noter` binary |
| `make install` | Build and create symlink in `~/.local/bin` |
| `make clean` | Remove binary and symlink |
| `make help` | Show available commands |

## Configuration

Create a `.env` file in the project root (or where you run the CLI):

```env
WORK_NOTES_LOCATION=~/Documents/obsidian-notes/Inbox/work
WORKPLACE_NAME=YourCompany
OPENCODE_SERVER=http://127.0.0.1:4096
AI_PROVIDER=github-copilot
AI_MODEL=claude-sonnet-4
```

| Variable | Description | Default |
|----------|-------------|---------|
| `WORK_NOTES_LOCATION` | Path to your Obsidian notes folder | `~/Documents/obsidian-notes/Inbox/work` |
| `WORKPLACE_NAME` | Name of your workplace (used in filenames and tags) | `Work` |
| `OPENCODE_SERVER` | URL of your OpenCode server for AI summaries | `http://127.0.0.1:4096` |
| `AI_PROVIDER` | AI provider ID for summaries | `github-copilot` |
| `AI_MODEL` | AI model ID for summaries | `claude-sonnet-4` |

## CLI Commands

### `noter start`

**Main command** - Start your daily workflow. This command:

1. Reviews pending items from the most recent previous note
2. Asks if each pending item was completed (y/n)
3. Moves completed items to yesterday's "Work Completed" section
4. Carries forward incomplete items to today's "Pending Work"
5. Generates an AI summary of yesterday's completed work
6. Creates today's note with the summary

```bash
noter start
```

### `noter add "task"`

Add a new pending work item to today's note.

```bash
noter add "Fix the login bug"
noter add "Review PR #123"
noter add "Update documentation for API endpoints"
```

### `noter done`

Interactively mark pending items as completed. Shows each pending item and asks if it's done.

```bash
noter done
```

### `noter list`

Display all pending and completed work items from today's note.

```bash
noter list
```

### `noter review`

Manually review pending items from previous notes without creating a new note or generating summaries.

```bash
noter review
```

### `noter summarize`

Generate and display an AI-powered summary of today's completed work (output only, does not save to file).

```bash
noter summarize
```

## Note Format

Notes are created with the filename format: `YYYY-MM-DD-WorkplaceName.md`

Example: `2025-01-19-Jio.md`

```markdown
---
id: Jio-19-Jan-2025
aliases: []
tags:
  - jio
  - job
date: 2025-01-19
---

# 2025-01-19

summary:: Resolved critical authentication issues and completed code review tasks.

yesterday's summary:: Fixed database connection issues and deployed hotfix to production.

## Pending Work

- [ ] Update API documentation
- [ ] Review team's pull requests

## Work Completed

- [x] Fix login authentication bug
- [x] Deploy v2.1.0 to staging
```

## Daily Workflow

### Morning Routine

1. Run `noter start` at the beginning of your workday
2. Review each pending item from yesterday:
   - Press `y` if completed
   - Press `n` to carry forward to today
3. The CLI generates an AI summary of yesterday's work
4. A new note is created for today

### During the Day

- Use `noter add "task"` to add new work items
- Use `noter done` to mark items as completed
- Use `noter list` to see your current progress

### Example Session

```
$ noter start
Starting daily workflow for 2025-01-19...

Found previous note: 2025-01-18-Jio.md (Date: 2025-01-18)

Reviewing pending items from previous note...

Did you complete: "Fix the login bug"? [y/N]: y
Did you complete: "Update documentation"? [y/N]: n

Generating AI summary of completed work...

Summary: Fixed critical login authentication bug affecting user sessions.

Updated previous note: 2025-01-18-Jio.md
Saved today's note: 2025-01-19-Jio.md

--- Today's Note ---

--- Pending Work ---
  1. [ ] Update documentation

--- Completed Work ---
  No completed items

Daily workflow complete! Use 'noter add "task"' to add new items.
```

## Requirements

- Go 1.21 or later
- OpenCode server running (for AI summaries)

## Dependencies

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [joho/godotenv](https://github.com/joho/godotenv) - Environment variable loading
- [manifoldco/promptui](https://github.com/manifoldco/promptui) - Interactive prompts

## License

MIT
