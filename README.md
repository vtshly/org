# Org

A simple terminal-based Org-mode task manager inspired by the simplicity of `nano`. Manage your TODO items, track time, and stay organized without leaving the command line.

## Installation

```bash
go install github.com/rwejlgaard/org/cmd/org@latest
```

Or build from source:
```bash
git clone https://github.com/rwejlgaard/org
cd org
go build -o bin/org ./cmd/org
```

## Usage

```bash
org [filename]           # Open specific org file
org -f tasks.org         # Open using -f flag
org                      # Opens ./todo.org by default
```

## Features

### Task Management
- **TODO States**: Cycle through TODO, PROG (in progress), BLOCK (blocked), and DONE states
- **Hierarchical Tasks**: Create sub-tasks and organize items with multiple levels
- **Priority Levels**: Set priorities (A, B, C) with color-coded indicators
- **Folding**: Collapse and expand tasks and notes with Tab key
- **Quick Capture**: Press 'c' to quickly capture new TODO items
- **Reorder Mode**: Reorganize tasks with shift+up/down arrows

### Scheduling & Deadlines
- **Deadlines**: Set and track task deadlines with visual indicators
- **Scheduled Dates**: Schedule tasks for specific dates
- **Agenda View**: View upcoming tasks for the next 7 days
- **Overdue Highlighting**: Automatically highlights overdue items in red

### Time Tracking
- **Clock In/Out**: Track time spent on tasks with 'i' (clock in) and 'o' (clock out)
- **Duration Display**: See current and total time tracked per task
- **Effort Estimates**: Set estimated effort (e.g., 8h, 2d, 1w)
- **Automatic Logging**: All clock entries are logged in LOGBOOK drawer

### Notes & Documentation
- **Rich Notes**: Add detailed notes to any task with Enter key
- **Syntax Highlighting**: Code blocks are automatically highlighted (supports both ```lang and #+BEGIN_SRC formats)
- **Markdown Support**: Use markdown-style code blocks in your notes
- **Drawer Management**: LOGBOOK and PROPERTIES drawers are automatically filtered in list view

### Keybindings

| Key | Action |
|-----|--------|
| `↑/k`, `↓/j` | Navigate up/down |
| `←/h`, `→/l` | Cycle state backward/forward |
| `t` or `space` | Cycle TODO state |
| `tab` | Fold/unfold item |
| `enter` | Edit notes |
| `c` | Capture new TODO |
| `s` | Add sub-task |
| `D` | Delete item (with confirmation) |
| `a` | Toggle agenda view |
| `i` | Clock in |
| `o` | Clock out |
| `d` | Set deadline |
| `p` | Set priority |
| `e` | Set effort |
| `r` | Toggle reorder mode |
| `shift+↑/↓` | Move item up/down |
| `ctrl+s` | Save |
| `?` | Toggle help |
| `q` or `ctrl+c` | Quit |

### Auto-save
Changes are automatically saved when you quit the application.

## Screenshots

### List view
![list view](./.imgs/list_view.png)

### Editing notes
![editing](./.imgs/editing.png)

### Prompts
![capture](./.imgs/capture_prompt.png)
![delete](./.imgs/delete_prompt.png)
![priority](./.imgs/priority_prompt.png)

## File Format

The application uses standard Org-mode file format (.org), making it compatible with Emacs Org-mode and other Org-mode tools.

## License

MIT