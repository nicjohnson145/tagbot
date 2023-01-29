# tagbot
Automatically created tags based on conventional commits

# Using manually

Download the relevant binary from [here](https://github.com/nicjohnson145/tagbot/releases/latest)
and place it in your `$PATH`. If you've cloned the repo via ssh, then just run `tagbot` from within
the repo you wish to create tags for. If you've cloned the repo via https then you'll need to export
`AUTH_TOKEN` as an access token with the ability to create tags.

# Using as a Github Action

TagBot can be ran locally, or through Github Actions. Below is an example setup to only create tags
when pushing to the default branch

```yaml
on:
  push:
    branches:
    - main
    tags-ignore:
    - '**'

jobs:
  build-tag:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v3
      with:
        fetch-depth: 0
    - name: TagBot
      uses: nicjohnson145/tagbot@latest
      id: tagbot
      env:
        AUTH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
### Note about triggering other workflows

The default `${{ secrets.GITHUB_TOKEN }}` [can't create additional workflows](https://github.com/orgs/community/discussions/27028#discussioncomment-3254360).
If you want to use tagbot to create new tags when code is pushed to main, and goreleaser to create
releases when a new tag is created (the whole reason I wrote tagbot :)) then you'll need to replace
the token with a users access token.

# Using commit-msg git hooks

Tagbot has commit-msg git hook functionality as well. To use this functionality place the following
script in your `.git/hooks` directory named `commit-msg` after downloading tagbot and adding it to
your `$PATH`

```sh
#! /usr/bin/env bash

tagbot commit-msg $1
```

### Global git hooks & disabling

Setting `core.hooksPath` in your global gitconfig can allow you to run tagbot for every repo you
clone. This greatly cuts down on repeated setup, as well as lowers the change that the hook will be
forgotten on a new clone. However, not *every* repo needs to conform to tagbot. Tagbot can be
disabled for an individual repo by running

```sh
git config --add tagbot.disable true
```

in any repo that you wish tagbots `commit-msg` hook not to run

# Running on pull requests

Tagbot can retroactively validate commit messages on pull requests (if not everyone uses the
commit-msg hook). This can be accomplished with the following github action
```yaml
on:
  pull_request

jobs:
  check-commits:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v3
      with:
        fetch-depth: 0
    - name: TagBot
      uses: nicjohnson145/tagbot@latest
      args:
      - pull-request

```
# Options

Tagbot supports a number of options, either on the command line or through environment variables

| Command Line | Environment | Use |
| ------------ | ----------- | --- |
| `--debug` | `DEBUG` | Enable debug logging |
| `--latest` | `LATEST` | Maintain a `latest` tag in addition to the SemVer tags |
| `--always-patch` | `ALWAYS_PATCH` | If the run were to result in no tag being created, instead create a tag with a patch version bump|
| `--remote-name` | `REMOTE_NAME` | Name of the remote to push tags to, defaults to `origin` |
| `--auth-method` | `AUTH_METHOD` | What method to use to auth, defaults to clone method of remote |
| `--auth-token` | `AUTH_TOKEN` | Token to use during HTTPS authentication |
| `--auth-key-path` | `AUTH_KEY_PATH` | Path to key to use during SSH authentication |
| `--base-branch` | `BASE_BRANCH` | Base branch for merge request, will attempt to infer from well known CI systems variables |
