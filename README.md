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
clone. This greatly cuts down on repeated setup, as well as lowers the chance that the hook will be
forgotten on a new clone. However, not *every* repo needs to conform to tagbot. Tagbot can be
disabled for an individual repo by running

```sh
git config --add tagbot.disable true
```

in any repo that you wish tagbots `commit-msg` hook not to run

# Options

Tagbot supports a number of options, which can be set in various methods detailed below

| Command Line | Environment Variable | Monorepo Config | Use |
| ------------ | -------------------- | --------------- | --- |
| `--log-level` | `LOG_LEVEL` | _not applicable_ | Set the logging verbosity |
| `--remote-name` | `REMOTE_NAME` | _not applicable_ | Override the remote tags will be pushed to |
| `--auth-method` | `AUTH_METHOD` | _not applicable_ | What method to use to auth, defaults to clone method of remote |
| `--auth-token` | `AUTH_TOKEN` | _not applicable_ |  Token to use during HTTPS authentication |
| `--auth-token-username` | `AUTH_TOKEN_USERNAME` | _not applicable_ |  Username to use during HTTPS authentication |
| `--auth-key-path` | `AUTH_KEY_PATH` | _not applicable_ |  Path to key to use during SSH authentication |
| `--monorepo` | `MONOREPO` | _not applicable_ | Execute tagbot in monorepo mode, maintaining multiple tags |
| `--monorepo-config-path` | `MONOREPO_CONFIG_PATH` | _not applicable_ | Override the default configuration file path |
| `--maintain-latest` | `MAINTAIN_LATEST` | `maintain-latest` | Indicates a "latest" tag should be maintained in addition to semver |
| `--latest-name` | `LATEST_NAME` | `latest-name` | Override the name of the "latest" tag, if maintained |
| `--no-v` | `NO_V` | `no-v` | Do not add a `v` prefix to tags |
| `--always-patch` | `ALWAYS_PATCH` | `always-patch` | If a commit were to trigger no tag being made, instead create a patch tag. Note: in monorepo mode, a commit must be _relevant_ to a component for this behavior to trigger |

# MonoRepos

Tagbot supports multiple "projects" within a single git repository. Each one can be tagged independently. This behavior
is controlled via 2 pieces of configuration: the monorepo flag, and the monorepo configuration file (default location of
`.tagbot.yaml` at repo root). Components are defined manually, with the changeset globs functioning to gate which
changed files should correspond to which components. Configuration supplied via the environment or command line flags
for certain settings set the "default" for all components, namely: `--maintain-latest`, `--latest-name`, `--no-v`, &
`--always-patch`. If these settings are set, and a component does not override them, the value passed there will be
used. See below for an example configuration:

```yaml
components:
  foo:
    change-set-globs:
    - src/pkg/foo/*
    - package.json
  bar:
    change-set-globs:
    - src/pkg/bar/*
    - package.json
    - src/libs/barlib/*
    maintain-latest: true
    latest-name: main
    no-v: true
    always-patch: true
```