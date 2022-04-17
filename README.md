# Bitbucket Semantic Pull Requests

Bitbucket Cloud status check that ensures your pull requests follow the [Conventional Commits spec](https://conventionalcommits.org/).

Heavily inspired by [zeke/semantic-pull-requests](https://github.com/zeke/semantic-pull-requests), if not a rewrite in Go for Bitbucket Cloud.

## How it works

By default, only the PR title OR at least one commit message needs to have semantic prefix. If you wish to change this
behavior, see [configuration](#configuration) section below.

Scenario | Status | Status Check Message
-------- | ------ | -------
PR title is semantic | 💚 | `ready to be squashed`
any commit is semantic | 💚 | `ready to be merged or fast-forwarded`
nothing is semantic | 💛 | `add a semantic commit or PR title`

Please see [zeke/semantic-pull-requests](https://github.com/zeke/semantic-pull-requests#how-it-works) for the full rational.

## Installation

1. Create a Bitbucket Cloud account for the bot and add it to your team (Recommended)
2. [Create an App password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) with [`pullrequest`](https://developer.atlassian.com/cloud/bitbucket/bitbucket-cloud-rest-api-scopes/#pullrequest) scope ("Pull request: Read")
3. [Grant read access on your repositories](https://support.atlassian.com/bitbucket-cloud/docs/grant-repository-access-to-users-and-groups/) to the bot account
4. Export the bot's Bitbucket credentials via environment variables (or use `--help` flag for more options):
    - `BITBUCKET_USERNAME`: Bitbucket username associated with the account used for renovate-approve-bot
    - `BITBUCKET_PASSWORD`: Bitbucket App password created in step 2
5. Start the bot:

     - With Docker:

       ```shell
       docker run --rm \
         --env BITBUCKET_USERNAME \
         --env BITBUCKET_PASSWORD \
         --publish 8888:8888 \
         ghcr.io/maxbrunet/bitbucket-semantic-pull-requests:latest
       ```

     - With a binary (downloadable from the [Releases](https://github.com/maxbrunet/bitbucket-semantic-pull-requests/releases) page):

       ```shell
       ./bitbucket-semantic-pull-requests
       ```

6. [Add a webhook to your repository](https://support.atlassian.com/bitbucket-cloud/docs/manage-webhooks/#Create-webhooks)

    - *Title*: Semantic Pull Requests
    - *URL*: `https://<bot-address>/`
    - *Status*: Active
    - *Triggers*:
        - `pullrequest:created`
        - `pullrequest:updated`


## Configuration

It is the same as [zeke/semantic-pull-requests](https://github.com/zeke/semantic-pull-requests).

By default, no configuration is necessary.

If you wish to override some behaviors, you can add a `semantic.yml` file to your `.bitbucket` directory with
the following optional settings:

```yml
# Disable validation, and skip status check creation
enabled: false
```

```yml
# Always validate the PR title, and ignore the commits
titleOnly: true
```

```yml
# Always validate all commits, and ignore the PR title
commitsOnly: true
```

```yml
# Always validate the PR title AND all the commits
titleAndCommits: true
```

```yml
# Require at least one commit to be valid
# this is only relevant when using commitsOnly: true or titleAndCommits: true,
# which validate all commits by default
anyCommit: true
```

```yml
# You can define a list of valid scopes
scopes:
  - scope1
  - scope2
  ...
```

```yml
# By default conventional types as definited by go-conventionalcommits are used.
# See "conventional": https://github.com/leodido/go-conventionalcommits#types
# You can override the valid types
types:
  - feat
  - fix
  - docs
  - style
  - refactor
  - perf
  - test
  - build
  - ci
  - chore
  - revert
```

```yml
# Allow use of Merge commits (e.g. "Merge branch 'master' into feature/ride-unicorns")
allowMergeCommits: true
```

```yml
# Allow use of Revert commits (e.g. "Revert "feat: ride unicorns"")
allowRevertCommits: true
```

## Note about conventional-changelog

The [`mergePattern`](https://github.com/conventional-changelog/conventional-changelog/tree/master/packages/conventional-commits-parser#mergepattern) parser option can be used to extract the Pull Request title from Bitbucket's merge message:

```yaml
parserOpts:
  mergePattern: '^Merged in (\S+) \(pull request #(\d+)\)$'
  mergeCorrespondence: ['branch', 'prId']
```

## License

[Apache 2.0](LICENSE)
