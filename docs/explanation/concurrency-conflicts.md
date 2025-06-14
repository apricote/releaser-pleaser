# Concurrency and Conflicts

## Why

`releaser-pleaser` works on the "shared global state" that is your project on GitHub/GitLab. Each execution reads from that state and makes changes to it. While `releaser-pleaser` is generally [idempotent](https://en.wikipedia.org/wiki/Idempotence), we still need to consider concurrent executions for two reasons: avoiding conflicts and saving resources.

### Avoiding conflicts

The [Release Pull Request](release-pr.md) is used by `releaser-pleaser` to show the current release. Users may update the PR description to add additional notes into the Changelog.

When `releaser-pleaser` is running while the user modifies the Release Pull Request description, `releaser-pleaser` may overwrite the description afterward based on its outdated local copy of the pull request.

### Saving resources

While `releaser-pleaser` is idempotent, there is no benefit to running it multiple times in parallel. In the best case, `releaser-pleaser` could be stopped as soon as a new "change" that is relevant to it comes in and restarts based on that new state.

## Measures taken

### Concurrency limits in CI environments

Our default configurations for [GitHub Actions](../tutorials/github.md) and [GitLab CI/CD](../tutorials/gitlab.md) try to limit concurrent `releaser-pleaser` jobs to a single one.

#### GitHub Actions

On GitHub Actions, we use a `concurrency.group` to restrict it to a single running job per repository.

GitHub cancels the currently running job and any other pending ones when a new one is started. This makes sure that `releaser-pleaser` always works with the latest state.

Users need to enable this in their workflow (included in our GitHub tutorial):

```yaml
concurrency:
  group: releaser-pleaser
  cancel-in-progress: true
```

#### GitLab

On GitLab CI/CD, we use a `resource_group: releaser-pleaser` in our GitLab CI/CD component to restrict it to a single running job per repository. This is part of the component YAML, so users do not need to set this manually.

There is no easy way to cancel the running job, so we let it proceed and rely on the other measures to safely handle the data. Users can enable "auto-cancel redundant pipelines" if they want, but should consider the ramifications for the rest of their CI carefully before doing so.

### Graceful shutdown

When GitHub Actions and GitLab CI/CD cancel jobs, they first sent a signal to the running process (`SIGINT` on GitHub and `SIGTERM` on GitLab). We listen for these signals and initiate a shutdown of the process. This helps save resources by shutting down as fast as possible, but in a controlled manner.

### Re-checking PR description for conflict

When `releaser-pleaser` prepares the Release Pull Request, the first step is to check if there is an existing PR already opened. It then reads from this PR to learn if the user modified the release in some way ([Release Notes](../guides/release-notes.md#for-the-release), [Pre-releases](../guides/pre-releases.md)). Based on this, it prepares the commit and the next iteration of the Release Pull Request description. The last step is to update the Release Pull Request description.

Depending on the time since the last release, a lot of API calls are made to learn about these changes; this can take between a few seconds and a few minutes. If the user makes any changes to the Release Pull Request in this time frame, they are not considered for the next iteration of the description. To make sure that we do not lose these changes, `releaser-pleaser` fetches the Release Pull Request description again right before updating it. In case it changed from the start of the process, the attempt is aborted, and the whole process is retried two times.

This does not fully eliminate the potential for data loss, but reduces the time frame from multiple seconds (up to minutes) to a few hundred milliseconds.

## Related Documentation

- **Explanation**
  - [Release Pull Request](release-pr.md)
- **Guide**
  - [Pre-releases](../guides/pre-releases.md)
  - [Customizing Release Notes](../guides/release-notes.md)
- **Tutorial**
  - [Getting started on GitHub](../tutorials/github.md)
  - [Getting started on GitLab](../tutorials/gitlab.md)

