# tagbot
Automatically created tags based on conventional commits

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
