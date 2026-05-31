# grons

A terminal UI for managing cron jobs. Browse, add, edit, delete, and toggle cron entries from a single interface. Shows run history pulled from `journalctl`.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Requirements

- Linux
- `crond` running
- `systemd` (for run history via `journalctl`)

## Installation

### Ubuntu / Debian

Download the `.deb` from the [latest release](https://github.com/gavasc/gronma/releases/latest):

```bash
wget https://github.com/gavasc/gronma/releases/latest/download/gronma_<version>_linux_amd64.deb
sudo dpkg -i gronma_<version>_linux_amd64.deb
```

For ARM64 (e.g. Raspberry Pi, ARM servers), replace `amd64` with `arm64`.

### Fedora / RHEL / CentOS

```bash
wget https://github.com/gavasc/gronma/releases/latest/download/gronma_<version>_linux_amd64.rpm
sudo rpm -i gronma_<version>_linux_amd64.rpm
```

### Alpine

```bash
wget https://github.com/gavasc/gronma/releases/latest/download/gronma_<version>_linux_amd64.apk
sudo apk add --allow-untrusted gronma_<version>_linux_amd64.apk
```

### Any Linux (binary)

```bash
wget https://github.com/gavasc/gronma/releases/latest/download/gronma_<version>_linux_amd64.tar.gz
tar -xzf gronma_<version>_linux_amd64.tar.gz
sudo mv gronma /usr/local/bin/
```

### Via `go install`

Requires Go 1.23+:

```bash
go install github.com/gavasc/gronma@latest
```

### Build from source

```bash
git clone https://github.com/gavasc/gronma.git
cd gronma
go build -o gronma .
sudo mv gronma /usr/local/bin/
```

## Usage

```bash
gronma
```

### Key bindings

#### List view

| Key     | Action               |
|---------|----------------------|
| `j`/`k` | Move up/down         |
| `Enter` | Open detail view     |
| `a`     | Add new entry        |
| `e`     | Edit selected entry  |
| `d`     | Delete selected entry|
| `Space` | Enable/disable entry |
| `q`     | Quit                 |

#### Editor

| Key       | Action              |
|-----------|---------------------|
| `Tab`     | Switch field        |
| `Ctrl-F`  | Open file picker    |
| `Ctrl-S`  | Save                |
| `Esc`     | Cancel              |

## What it reads

- **User crontab** — `crontab -l` (editable)
- **System crontabs** — `/etc/crontab` and `/etc/cron.d/*` (read-only)
- **Run history** — `journalctl -u crond` (last 50 runs per entry)

Disabled entries are stored with the prefix `##CRONMA_DISABLED## `, compatible with [cronma](https://github.com/gavasc/cronma).
