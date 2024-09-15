# Updating arbitrary files

In some situations it makes sense to have the current version committed in files in the repository:

- Documentation examples
- A source-code file that has the version for user agents and introspection
- Reference to a container image tag that is built from the repository

`releaser-pleaser` can automatically update these references in the [Release PR](../explanation/release-pr.md).

## Markers

The line that needs to be updated must have the marker `x-releaser-pleaser-version` somewhere after the version that should be updated.

For example:

```go
// version/version.go

package version

const Version = "v1.0.0" // x-releaser-pleaser-version
```

## Extra Files

You need to tell `releaser-pleaser` which files it should update. This happens through the CI-specific configuration.

### GitHub Action

In the GitHub Action you can set the `extra-files` input with a list of the files. They need to be formatted as a single multi-line string with one file path per line:

```yaml
jobs:
  releaser-pleaser:
    steps:
      - uses: apricote/releaser-pleaser@v0.4.0
        with:
          extra-files: |
            version.txt
            version/version.go
            docker-compose.yml
```

### GitLab CI/CD Component

In the GitLab CI/CD Component you can set the `extra-files` input with a list of files. They need to be formatted as a single multi-line string with one file path per line:

```yaml
include:
  - component: $CI_SERVER_FQDN/apricote/releaser-pleaser/run@v0.4.0
    inputs:
      extra-files: |
        version.txt
        version/version.go
        docker-compose.yml
```

## Related Documentation

- **Reference**
  - [GitHub Action](../reference/github-action.md#inputs)
  - [GitLab CI/CD Component](../reference/gitlab-cicd-component.md#inputs)
