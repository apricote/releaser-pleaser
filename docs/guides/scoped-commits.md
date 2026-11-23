# Using Scoped Commits

Many repositories do not use [Conventional Commits](../reference/glossary.md#conventional-commits). Instead, they follow [Scoped Commits](https://scopedcommits.com/): each commit message starts with the part of the project that it changes, for example `api: add endpoint` or `ui: fix button color`. `releaser-pleaser` supports this format with the [`scoped` commit parser](../reference/commit-parsers.md#scoped-commits).

## Choosing a versioning scheme

Scoped Commits do not say whether a change is a feature, a fix or a breaking change, so `releaser-pleaser` can not decide which part of a [Semantic Version](../reference/glossary.md#semantic-versioning-semver) to bump. The `scoped` commit parser can therefore only be used with [Calendar Versioning](calendar-versioning.md), where the version is built from the release date.

## Configuration

Set the `commit-parser` and `versioning` inputs.

GitHub Actions:

```yaml
- uses: apricote/releaser-pleaser@v0.9.0
  with:
    commit-parser: scoped
    versioning: calver
```

GitLab CI/CD:

```yaml
include:
  - component: $CI_SERVER_FQDN/apricote/releaser-pleaser/run@v0.9.0
    inputs:
      token: $RELEASER_PLEASER_TOKEN
      commit-parser: scoped
      versioning: calver
```

To match the commit format, the release pull request and the release commit are titled `release: <version>`, for example `release: 2026.47.0`.

## Release Notes

Every commit since the last release is listed in a single `Changes` section of the release notes. Only the first line of the commit message is used. Commits that do not start with a scope are listed without one:

```markdown
### Changes

- **api**: add endpoint ([abc1234](https://example.com/commit/abc1234))
- fix typo in readme ([def5678](https://example.com/commit/def5678))
```

Merge commits with the default message of git or your forge, like `Merge pull request #12 from owner/foo`, are not listed, but the commits of the merged branch are. If you prefer a single entry for each pull request, squash pull requests when merging them.

## Editing the Release Notes

As every commit is listed, you might want to change or remove some entries. The `rp-commits` code block in the pull request description works with the `scoped` commit parser as well, see [Customizing Release Notes](release-notes.md#editing-the-release-notes):

> ```rp-commits
> api: add movie endpoints
> api: add cinema endpoints
> ```

## Related Documentation

- **Guides**
  - [Calendar Versioning](calendar-versioning.md)
  - [Customizing Release Notes](release-notes.md)
- **Reference**
  - [Commit Parsers](../reference/commit-parsers.md)
  - [Versioning](../reference/versioning.md)
  - [GitHub Action](../reference/github-action.md)
  - [GitLab CI/CD Component](../reference/gitlab-cicd-component.md)
