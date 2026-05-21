# OE — Terminal Text Editor

> A cross-platform terminal text editor inspired by the classic MS-DOS `edit` command, built with Go and tcell.

**OE** (Open Editor) 是一款跨平台终端文本编辑器，灵感来自经典的 Windows MS-DOS `edit` 命令。提供 View/Edit 双模式、命令面板、文件浏览器等现代化功能。

---

## Features / 功能特性

| English | 中文 |
|---------|------|
| View/Edit dual mode (read-only by default) | View/Edit 双模式（默认只读） |
| Command palette (`Ctrl+P`) with fuzzy search | 命令面板 (`Ctrl+P`)，支持模糊搜索 |
| File browser with directory navigation (`Ctrl+O`) | 文件浏览器，支持目录导航 (`Ctrl+O`) |
| Mouse support (click, scroll, drag-select) | 鼠标支持（点击、滚轮、拖拽选择） |
| System clipboard (copy/cut/paste) | 系统剪贴板（复制/剪切/粘贴） |
| Search (`Ctrl+F`) with find-next (`F3`) | 搜索 (`Ctrl+F`)，查找下一个 (`F3`) |
| Three-way save dialog (Save/Don't Save/Cancel) | 三选保存对话框（保存/不保存/取消） |
| Undo (snapshot-based, up to 100 steps) | 撤销（基于快照，最多 100 步） |
| Line numbers | 行号显示 |
| Status bar with mode indicator | 状态栏（模式/文件名/光标位置） |
| Full CJK character support | 完整中文/日文/韩文支持 |
| Cross-platform (Linux, macOS, Windows) | 跨平台（Linux / macOS / Windows） |

---

## Screenshot / 界面预览

```
┌─────────────────────────────────────────────────┐
│  /home/user/projects                             │
│  Path> test.txt                             █    │
│─────────────────────────────────────────────────│
│  ↑ [up-dir]                                     │
│  > src                                           │
│  > docs                                          │
│    main.go                                       │
│    go.mod                                        │
│    README.md                                     │
│    test.txt                                      │
├─────────────────────────────────────────────────┤
│ [VIEW]  main.go  42:8  UTF-8    Ctrl+P:命令面板  │
└─────────────────────────────────────────────────┘
```

---

## Quick Start / 快速开始

```bash
# Clone / download
git clone https://github.com/zsying/oe.git
cd oe

# Build
go build -ldflags="-s -w" -o oe .

# Run
./oe               # New blank file
./oe README.md     # Open existing file
```

### Dependencies / 依赖

- Go 1.21+
- [tcell v2](https://github.com/gdamore/tcell/v2) — terminal UI
- [go-runewidth](https://github.com/mattn/go-runewidth) — wide character support
- [atotto/clipboard](https://github.com/atotto/clipboard) — system clipboard

---

## Keybindings / 快捷键

### Global / 全局

| Key | Action |
|-----|--------|
| `Ctrl+E` | Toggle View/Edit mode |
| `Ctrl+P` | Command palette |
| `Ctrl+O` | Open file (file browser) |
| `Ctrl+S` | Save file |
| `Ctrl+Q` | Quit |
| `Ctrl+F` | Find |
| `F3` | Find next |
| `Ctrl+H` | Replace |
| `Ctrl+Z` | Undo |

### Navigation / 导航

| Key | Action |
|-----|--------|
| `↑↓←→` | Move cursor |
| `PgUp` / `PgDn` | Page up / down |
| `Home` / `End` | Start / end of line |

### Edit Mode / 编辑模式

| Key | Action |
|-----|--------|
| `Ctrl+X` | Cut |
| `Ctrl+C` | Copy |
| `Ctrl+V` | Paste |
| `Ctrl+A` | Select all |
| `Del` | Delete forward |
| `Backspace` | Delete backward |
| `Enter` | New line |
| `Tab` | Insert tab (4 spaces) |

### File Browser / 文件浏览器

| Key | Action |
|-----|--------|
| `Typing` | Enter path (focus on input) |
| `↓ / ↑` | Switch to file list / navigate |
| `Tab` | Toggle focus (input ↔ list) |
| `Enter` | Confirm path / open file |
| `Esc` | Cancel |

---

## Project Structure / 项目结构

```
oe/
├── main.go              # Entry point
├── internal/
│   ├── buffer/          # Text buffer (line array)
│   ├── editor/          # Editor state & commands
│   ├── screen/          # Terminal UI (tcell)
│   ├── widgets/         # UI widgets (status bar, dialog, etc.)
│   └── clipboard/       # System clipboard wrapper
├── docs/                # Design documents & plans
└── go.mod
```

---

## License / 许可

MIT
