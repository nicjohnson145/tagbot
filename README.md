# tagbot
Automatically created tags based on conventional commits

# Using manually

Download the relevant binary from [here](https://github.com/nicjohnson145/tagbot/releases/latest)
and place it in your `$PATH`. If you've cloned the repo via ssh, then just run `tagbot` from within
the repo you wish to create tags for. If you've cloned the repo via https then you'll need to export
`TAGBOT_TOKEN` as an access token with the ability to create tags.

# Using as a Github Action

TagBot can be ran locally, or through Github Actions. Below is an example setup to only create tags
when pushing to the default branch

```yaml
on:
  push:
    branches:
    - main

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
        TAGBOT_TOKEN: ${{ secrets.GITHUB_TOKEN }}
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
      env:
        TAGBOT_TOKEN: ${{ secrets.GITHUB_TOKEN }}

```
