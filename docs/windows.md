# lnk on Windows

`lnk` runs natively on Windows. It manages files in your Windows home directory and makes native symlinks. This page lists the requirements, the install steps and the differences from Linux and macOS.

The Windows support is in the fork `wolffshots/lnk`. A release from `yarlson/lnk` does not include it. Tested on Windows 11 with PowerShell.

## Requirements

- Git for Windows, with `git` on the `PATH`.
- The right to make symlinks. Turn on Developer Mode in Settings, System, For developers. An elevated shell also works.
- Go 1.25 or later, to build from source.

Without the symlink right, `lnk` changes nothing and prints "Windows did not allow lnk to create a symlink".

## Install

Build from source. Run these commands in PowerShell.

```powershell
git clone https://github.com/wolffshots/lnk.git
cd lnk
go build -o lnk.exe .
```

Move `lnk.exe` to a directory on your `PATH`.

The `install.sh` script and Homebrew do not support Windows. The fork has no release yet.

## First use

The repo path is `%USERPROFILE%\.config\lnk`. Set `LNK_HOME` to use a different directory on the same drive.

```powershell
lnk init
lnk add --host mypc $HOME\.gitconfig
lnk add --host mypc $HOME\.agents\skills\*
lnk list --host mypc
```

- Use `$HOME` in PowerShell for the home directory.
- `lnk add` expands `*` and `?` itself, because PowerShell and `cmd.exe` do not. `lnk rm` takes one path and no wildcard.
- `lnk` shows paths with backslashes. The `.lnk` index files use forward slashes on every operating system, so one repo works on Windows, Linux and macOS.

## Use a repo from Linux or macOS

```powershell
lnk init -r git@github.com:you/dotfiles.git
lnk pull --host mypc
```

Be careful with `lnk pull` when you give no `--host`. It links each common entry into your Windows home directory. A repo made for Linux can hold `.zshrc` or `.gitconfig` there. If a real file is in the way, `lnk` renames it to `<name>.lnk-backup` first.

Put the Windows files under a host name, and pass `--host` to `add`, `rm`, `list`, `pull` and `doctor`. Use a host name that no other machine uses. A WSL distribution on the same PC is a different machine with a different home directory.

## Differences from Linux and macOS

| Subject | Behavior on Windows |
|---|---|
| Files outside the home directory | `lnk add` rejects them. |
| Repo path on a different drive | `lnk` rejects it. A relative symlink cannot cross drives. |
| `bootstrap.sh` | `lnk` does not run it. `lnk init -r` prints a line and continues. `lnk bootstrap` returns an error. Run the script yourself from Git Bash or WSL. |
| Wildcards | `lnk add` expands `*` and `?`. |
| File name case | Windows ignores case and the index file does not. Keep one spelling for each path. |

A path longer than 260 characters needs `core.longpaths` in the repo. Without it, `lnk add` on such a file fails with a Git error and moves nothing.

```powershell
git -C $HOME\.config\lnk config core.longpaths true
```

## Line endings

Git for Windows often sets `core.autocrlf` to `true`. Git then writes CRLF line endings into the stored files on Windows. The files in the Git history keep LF. If a tool needs LF, turn the conversion off for the repo:

```powershell
git -C $HOME\.config\lnk config core.autocrlf false
```

The setting applies to files that Git writes after the change. A `.gitattributes` file with the line `* -text` in the dotfiles repo does the same for every machine.

## Errors and fixes

| Message | Cause | Fix |
|---|---|---|
| Windows did not allow lnk to create a symlink | The shell has no symlink right. | Turn on Developer Mode, or run `lnk` as administrator. |
| The lnk repository and the file are on different drives or file systems | The repo path and the file are on different drives. | Set `LNK_HOME` to a directory on the drive of your home directory. |
| Cannot move the file because Windows denied access | A program has the file open, or a file inside the directory. | Close the program and run the command again. |
| Cannot manage a file outside the home directory on Windows | The path is not under `%USERPROFILE%`. | Move the file into the home directory, or do not manage it with `lnk`. |
| lnk does not run bootstrap scripts on Windows | You ran `lnk bootstrap`. | Run the script from Git Bash or WSL. |

When `lnk add` fails, it moves no file. When a restore fails, an existing file stays in place.

If the console shows broken characters in place of emoji, pass `--no-emoji`.

## WSL

A WSL distribution has its own home directory and its own `lnk` repo. Use the tool that matches the home directory:

- For files under `C:\Users\<name>`, run `lnk.exe` from PowerShell.
- For files in the WSL home directory, run the Linux `lnk` in WSL.

Do not run the Linux `lnk` on files under `/mnt/c`. The move fails with the "different drives or file systems" error. A symlink that WSL makes on the Windows drive can also be unreadable for Windows programs.
