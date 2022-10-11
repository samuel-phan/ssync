# ssync

Source sync helps you to do recurring copies of a directory to a remote host.

Typical use case is this:

You are working on project on personal computer.

- you **modify some files locally**.
- you want to **send the current sources to a remote host** (that may be on a
  cloud or in a protected network) for **testing** (compilation, execution in a
  specific environment that you can't have locally).

and you repeat this loop during your work.

In the essence, it is just a wrapper around `rsync`.

# Prerequisites

- `rsync` command-line tool.

# Installation

Go to the `Releases` page on the GitHub project and download the tarball for
your platform.

For example, for macOS (Darwin):

```bash
ssync_x.x.x_Darwin_x86_64.tar.gz
```

Unpack it:

```bash
tar -xvf ssync_x.x.x_Darwin_x86_64.tar.gz
```

Copy the `ssync` executable somewhere in your `$PATH`. For example:

```bash
$HOME/bin
```

If it is not in your `$PATH`, add it to your `~/.bash_profile` (macOS),
`~/.bashrc` (linux) or `~/.zshrc`:

```bash
export PATH="$PATH:$HOME/bin"
```

Reload your shell:

```bash
# If you use bash
source ~/.bash_profile  # on macOS
source ~/.bashrc  # on linux

# If you use zsh
source ~/.zshrc
```

# User guide

## Adhoc copy

You can run an adhoc copy to a remote server without any prior initialization:

```bash
ssync myserver.example.com:/dest/path
```

You will see the executed `rsync` command:

```bash
Running: rsync -avzP --exclude .idea --exclude .vscode --exclude .terraform --exclude *.tfstate.backup --exclude *.py[co] --exclude __pycache__ --exclude foo /my/current/working/foodir myserver.example.com:/dest/path
```

And on the remote server, you will have this directory:

```
myserver.example.com:/dest/path/foodir
```

The `--exclude` files come from your user configuration file (see
[User configuration file](#user-configuration-file)).

## With project initialization

If you plan to copy several times the same directory, you should initialize the
directory you want to copy. We call it the **project directory**.

To initialize a project directory:

```bash
# creates ".ssync" file in the current directory
ssync -init

# or creates "another/directory/.ssync" file
ssync -init another/directory
```

Here is the content of the `.ssync`:

```yaml
nodes:
- server:/path
excludes:
- .idea
- .vscode
- .terraform
- '*.tfstate.backup'
- '*.py[co]'
- __pycache__
delete: true
sudo-user: ""
extra-args: []
```

- `nodes`: the list of target nodes to copy to. The format is the same as the
  rsync's one, ie `SERVER:[remote/path]` (requires the `:`).
- `excludes`: the list of excluded files.
- `delete`: delete extraneous files from destination dirs.
- `sudo-user`: destination user ownership. If you want to copy the destination
  files under another username (you must be able to `sudo su` as this
  destination user). Typical use case is when you copy files to a remote server
  in a directory owned by another user (eg: `jenkins` or `root`).
- `extra-args`: extra arguments to pass to `rsync`.

Customize the `nodes`, and **copy the project directory to the remote nodes**:

```bash
ssync
```

You can call `ssync` **from any subdirectory of the project directory**.

For example:

- your project directory is `/my/project`
- your current working directory is `/my/project/lib/foo`

If you run `ssync` from the directory `/my/project/lib/foo`, `ssync` will
search for a `.ssync` in this order:

- `/my/project/lib/foo/.ssync`
- `/my/project/lib/.ssync`
- `/my/project/.ssync`
- `/my/.ssync`
- `/.ssync`

If no project directory is found, it will assume that this is an **adhoc call**.

## Pass arbitrary arguments to `rsync` directly

You can pass arbitrary arguments to `rsync` directly after the `--` argument.

For example:

```bash
ssync -n -- -q --exclude-from=my_excludes
```

- `-n` is evaluated by `ssync` as dry-run flag.
- `-q --exclude-from=my_excludes` is passed as is to `rsync`.

You can also add these arbitrary arguments in the `.ssync` file like so:

```yaml
extra-args:
- -n
- -q
- --exclude-from=my_excludes
```

or this to forward your ssh-agent keys:

```yaml
extra-args:
- -e
- ssh -A
```

## User configuration file

`ssync` will create a user configuration file `~/.config/ssync/config.yaml` on
the first run that is different from `-help`.

Here is an example:

```yaml
excludes:
- .idea
- .vscode
- .terraform
- '*.tfstate.backup'
- '*.py[co]'
- __pycache__
```

- `excludes`: file patterns that will be excluded from the rsync copy. This
  excludes list is used:
  - for adhoc copies (without any project initialization).
  - for project initialization (`ssync -init`).

# Troubleshooting

If you are facing any issue or have a suggestion, please feel free to file an
issue on the `Issues` page of the project.

# Developer guide

## Requirements

- Go 1.16+
- GoReleaser (https://github.com/goreleaser/goreleaser)

## Run locally

```bash
go run .
```

## Install locally

```bash
# build and copy the binary to your $GOPATH/bin
go install
```

## How to make a release

### Install GoReleaser

```
brew install goreleaser
```

### Generate a GitHub token

GoReleaser needs a new GitHub token to make a release in your repository and
upload the artifacts:

- Go to https://github.com/settings/tokens/new
- Select the `repo` scope.
- Click on `Generate token`.

### With `make-release.sh`

```bash
./make-release.sh
```

### Manual release

Export your `GITHUB_TOKEN`:

```bash
export GITHUB_TOKEN="YOUR_GH_TOKEN"
```

Test the GoRelease configuration:

```bash
goreleaser --snapshot --skip-publish --rm-dist
```

Make the real release:

```bash
# Make sure you pushed everything to the remote
git push

# Create a tag and push it to GitHub
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Check the GoRelease locally
goreleaser --skip-publish --rm-dist

# If all good, do the real release
goreleaser --rm-dist
```
