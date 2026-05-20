# CLI Editor — Design Document

> 类似 Windows `edit` 命令的跨平台终端文本编辑器，Go 语言实现

## 概述

基于 `tcell` 的 TUI 编辑器，支持 View/Edit 双模式，默认以只读视图启动。
核心交互通过命令面板 (`Ctrl+P`) 完成，辅以菜单栏和快捷键。

## 技术选型

| 层 | 选择 | 理由 |
|---|------|------|
| 终端库 | `tcell` v2 | 跨平台成熟、精确控制布局、支持鼠标事件 |
| 缓冲区 | 行数组 `[]string` | 标准版编辑器，中小文件性能足够，代码直观 |
| 剪贴板 | `atotto/clipboard` | 已覆盖 Win/Mac/Linux(X11+Wayland) |
| Go 版本 | 1.21+ | 跨平台编译支持 |

## 项目结构

```
editor/
├── main.go                 # CLI 入口、参数解析
├── go.mod / go.sum
├── internal/
│   ├── buffer/             # 文本缓冲区
│   │   ├── buffer.go       #   行数组存储、插入/删除/加载/保存
│   │   ├── cursor.go       #   光标移动（字符、单词、行首/尾）
│   │   └── selection.go    #   选区管理
│   ├── editor/             # 编辑器核心
│   │   ├── editor.go       #   主状态结构体
│   │   ├── command.go      #   命令注册表（供命令面板调用）
│   │   ├── file.go         #   文件 I/O
│   │   └── mode.go         #   View/Edit 模式定义与切换
│   ├── screen/             # 终端 I/O
│   │   ├── screen.go       #   tcell 初始化、事件循环
│   │   └── render.go       #   渲染管线（菜单→编辑区→状态栏）
│   ├── widgets/            # UI 控件
│   │   ├── menubar.go      #   顶部下拉菜单栏
│   │   ├── statusbar.go    #   底部状态栏
│   │   ├── commandpalette.go # 命令面板（Ctrl+P）
│   │   └── dialog.go       #   通用对话框（确认、输入）
│   └── clipboard/          # 剪贴板封装
│       └── clipboard.go    #   适配 atotto/clipboard
└── docs/
    └── design.md           # 本文档
```

## 架构与数据流

### 模块依赖关系

```
main
  └─ screen (初始化 tcell)
       └─ editor (持有 buffer)
            ├─ buffer (行数组)
            ├─ clipboard (系统剪贴板)
            └─ command (命令注册表)
       └─ widget (菜单/状态栏/命令面板/对话框)
```

### 事件循环

```
tcell.Event → screen (事件解析)
                ├─ 键盘事件 → 判断模式(View/Edit) → 执行命令 → 更新 buffer → render()
                ├─ 鼠标事件 → 点击/拖拽/滚轮 → 定位光标/选区 → render()
                └─ 窗口大小变化 → 重新计算布局 → render()
```

### 渲染管线 (render)

每帧绘制顺序：
1. `screen.Clear()`
2. `menubar.Render(screen, width)` — 顶部固定行
3. `renderer.RenderLines(screen, buffer, offset, width, height)` — 编辑内容（带行号）
4. `statusbar.Render(screen, width)` — 底部固定行
5. **如果有浮动层**: `commandpalette.Render(screen, width, height)` 或 `dialog.Render(screen, width, height)`
6. `screen.ShowCursor(x, y)`
7. `screen.Sync()`

## 模式系统

| 特性 | View 模式（默认） | Edit 模式 |
|------|-------------------|-----------|
| 光标导航 | ↑↓←→ / PgUp/Dn / Home/End | 同 View |
| 文本选择 | 鼠标拖拽 / Shift+方向键 | 同 View |
| 复制 | Ctrl+C | Ctrl+C |
| 剪切 | ❌ | Ctrl+X |
| 粘贴 | ❌ | Ctrl+V |
| 插入/删除 | ❌ | 全部字符 + Enter + Backspace |
| 模式切换 | Ctrl+E | Ctrl+E |
| 状态栏标识 | `[VIEW]` | `[EDIT]` |

**模式切换时的行为：**
- View → Edit：光标位置不变，状态栏切换显示
- Edit → View：保存修改，光标位置不变

## 命令系统

每个命令有唯一 ID、显示标题、可选快捷键、执行函数。

### 预置命令清单

| 类别 | 命令 | ID | 快捷键 |
|------|------|----|--------|
| File | New | `file.new` | |
| File | Open… | `file.open` | Ctrl+O |
| File | Save | `file.save` | Ctrl+S |
| File | Save As… | `file.saveAs` | |
| File | Exit | `app.quit` | Ctrl+Q |
| Edit | Undo | `edit.undo` | Ctrl+Z |
| Edit | Cut | `edit.cut` | Ctrl+X |
| Edit | Copy | `edit.copy` | Ctrl+C |
| Edit | Paste | `edit.paste` | Ctrl+V |
| Edit | Delete | `edit.delete` | Del |
| Edit | Select All | `edit.selectAll` | Ctrl+A |
| View | Toggle Line Numbers | `view.toggleLineNum` | |
| View | Toggle Mode | `mode.toggle` | Ctrl+E |
| Find | Find… | `find.find` | Ctrl+F |
| Find | Find Next | `find.next` | F3 |
| Find | Replace… | `find.replace` | Ctrl+H |

### 命令面板

- 触发：`Ctrl+P`（无论 View/Edit）
- UI：浮动于编辑区之上，顶部输入框 + 下方匹配列表
- 交互：输入文字 → fuzzy 过滤命令标题 → ↑↓ 选择 → 回车执行 / Esc 关闭
- 实现：视图层叠加，不修改 buffer

## 剪贴板

| 操作 | View | Edit |
|------|------|------|
| 复制 | Ctrl+C 写入系统剪贴板 | 同 View |
| 剪切 | ❌ | Ctrl+X 写入系统剪贴板 + 删除选区 |
| 粘贴 | ❌ | Ctrl+V 从系统剪贴板读取 + 插入 |

依赖库 `github.com/atotto/clipboard`，在初始化时检测可用性，不可用时在状态栏提示。
Linux 下依赖 xclip/xsel（X11）或 wl-clipboard（Wayland）。

## 菜单布局

```
┌─ File ────────┬─ Edit ────────┬─ Selection ──┬─ Find ────────┬─ Help ────┐
│ New           │ Undo    Ctrl+Z│ Select All   │ Find… Ctrl+F  │ About     │
│ Open… Ctrl+O  │───────────────│   Ctrl+A     │ Find Next F3  │           │
│───────────────│ Cut     Ctrl+X│              │ Replace…      │           │
│ Save   Ctrl+S │ Copy    Ctrl+C│              │   Ctrl+H      │           │
│ Save As…      │ Paste   Ctrl+V│              │               │           │
│───────────────│ Delete  Del   │              │               │           │
│ Exit   Ctrl+Q │───────────────│              │               │           │
│               │ View/Edit     │              │               │           │
│               │ Mode   Ctrl+E │              │               │           │
└───────────────┴───────────────┴──────────────┴───────────────┴───────────┘
```

## 状态栏格式

```
  [VIEW]  main.go  42:8  UTF-8  (Windows CRLF)       未保存 ●
```

从左到右：模式标识、文件名、行:列、编码、行尾格式、修改标记。

### Undo（撤销）

标准版采用**快照式撤销**：每次修改前对整行数组做一次浅拷贝快照，撤销时回退到上一个快照。

- 撤销栈：保存最近 100 个快照（内存上限约 1MB）
- Ctrl+Z：回退一步
- 修改后新操作自动清空重做栈
- 初始版本暂不实现重做（Ctrl+Y），后续可加

## 剪贴板兼容性

| 平台 | 依赖 | 备注 |
|------|------|------|
| Windows | 无额外依赖 | 原生 API |
| macOS | 无额外依赖 | pbpaste/pbcopy |
| Linux X11 | xclip 或 xsel | 运行时检测 |
| Linux Wayland | wl-clipboard | 运行时检测 |
| SSH/无 GUI    | 内置寄存器   | 编辑器内复制/粘贴可用，不依赖系统剪贴板 |

### 行尾处理

- 读取文件时自动检测行尾格式（LF / CRLF）
- macOS / Linux 新文件默认 LF
- Windows 新文件默认 CRLF
- 保存时保持原始行尾格式
| SSH/Sway | 提供 fallback：内置寄存器 | 仅编辑器内复制/粘贴 |

## 错误处理策略

- 文件无法打开 → 状态栏显示错误 3 秒，不退出
- 文件无法保存 → 弹出错误对话框，可选择另存为
- 剪贴板不可用 → 注册 fallback（内部寄存器），状态栏提示
- 终端过小（< 80×24）→ 提示并等待调整

## CLI 参数

```
editor [filename]
```

- 无参数：空白新文件
- 有参数：打开指定文件（不存在则创建新文件）
- 无额外 flag 设计，保持简单
