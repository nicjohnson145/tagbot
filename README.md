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
      uses: nicjohnson145/tagbot@v0.1.0
      id: tagbot
      env:
        TAGBOT_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
### Note about triggering other workflows

The default `${{ secrets.GITHUB_TOKEN }}` [can't create additional workflows](https://github.com/orgs/community/discussions/27028#discussioncomment-3254360).
If you want to use tagbot to create new tags when code is pushed to main, and goreleaser to create
releases when a new tag is created (the whole reason I wrote tagbot :)) then you'll need to replace
the token with a users access token.
