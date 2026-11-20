# Commit Parsers

The commit parser reads the messages of all commits since the last release. It decides which commits are releasable and how they are listed in the release notes.

You can configure the commit parser through the `commit-parser` input on GitHub Actions and GitLab CI/CD, or the `--commit-parser` flag of `rp run`:

```
commit-parser: scoped
```

These rules apply to all commit parsers:

- Release commits created by `releaser-pleaser` are ignored.
- Merge commits with the default message of git or your forge are ignored, for example `Merge pull request #12 from owner/foo` or `Merge branch 'foo' into 'main'`. Merge commits with a custom message, like the title of the pull request, are parsed like any other commit.
- If the pull request of a commit has a `rp-commits` code block in its description, the lines of the code block are parsed instead of the commit message. See [Customizing Release Notes](../guides/release-notes.md#editing-the-release-notes).

## Conventional Commits

- **Name**: `conventional-commits`
- **Default**: yes
- **Example**: `feat(api): add endpoint`

Commit messages follow the [Conventional Commits](glossary.md#conventional-commits) specification. Only some commit types are releasable:

| Commit                          | Section in release notes |
| ------------------------------- | ------------------------ |
| `feat`                          | Features                 |
| `fix`                           | Bug Fixes                |
| Breaking change (`!` or footer) | marked as **BREAKING**   |

Breaking changes are marked with `!` after the type or scope (e.g. `feat(api)!: remove endpoint`) or with a `BREAKING CHANGE:` footer. Commits of other types, like `chore` or `docs`, and commits that do not follow the specification are not releasable and are not listed.

The commit types decide the next version with [Semantic Versioning](versioning.md#semantic-versioning).

## Scoped Commits

- **Name**: `scoped`
- **Default**: no
- **Example**: `api: add endpoint`

Commit messages follow the [Scoped Commits](https://scopedcommits.com/) format `<scope>: <description>`, where the scope is the part of the project that the commit changes.

The scope can not contain spaces or colons. Commits with multiple scopes (`api, ui: add endpoint`) or a ticket number after the scope (`auth (PROJ-123): fix login`) are listed with their whole first line and without a scope.

- Every commit is releasable.
- Only the first line of the commit message is used.
- Commits that do not start with a scope are listed with their whole first line and without a scope.

All commits are listed in a single `Changes` section of the release notes:

```markdown
### Changes

- **api**: add endpoint ([abc1234](https://example.com/commit/abc1234))
- fix typo in readme ([def5678](https://example.com/commit/def5678))
```

The commits do not tell which part of a version to bump, so this parser can only be used with [Calendar Versioning](versioning.md#calendar-versioning).

Learn more in the [Using Scoped Commits](../guides/scoped-commits.md) guide.
