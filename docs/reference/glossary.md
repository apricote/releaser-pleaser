# Glossary

### Changelog

The Changelog is a file in the repository (
`CHANGELOG.md`) that contains the [Release Notes](#release-notes) for every release of that repository. Usually, new releases are added at the top of the file.

### Conventional Commits

[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) is a specification for commit messages. It is the only supported commit message schema in
`releaser-pleaser`. Follow the link to learn more.

### Forge

A **forge
** is a web-based collaborative software platform for both developing and sharing computer applications.[^wp-forge]

Right now only **GitHub** is supported. We plan to support **GitLab
** in the future ([#4](https://github.com/apricote/releaser-pleaser/issues/4)). For other forges like Forgejo or Gitea, please open an issue and submit a pull request.

[^wp-forge]: Quote from [Wikipedia "Forge (software)"](<https://en.wikipedia.org/wiki/Forge_(software)>)

### Markdown

[Markdown](https://en.wikipedia.org/wiki/Markdown) is a lightweight markup language used on many [forges](#forge) as the preferred way to format text.

In `releaser-pleaser` Markdown is used for most texts.

### Pre-release

Pre-releases are a concept of [SemVer](#semantic-versioning-semver). They follow the normal versioning schema but use a suffix out of
`-alpha.X`, `-beta.X` and `-rc.X`.

Pre-releases are not considered "stable" and are usually not recommended for most users.

Learn more in the [Pre-releases](../guides/pre-releases.md) guide.

### Release Pull Request

A Release Pull Request is opened by
`releaser-pleaser` whenever it finds releasable commits in your project. It proposes a new version number and the Changelog. Once it is merged,
`releaser-pleaser` creates a matching release.

Learn more in the [Release Pull Request](../explanation/release-pr.md) explanation.

### Release Notes

Release Notes describe the changes made to the repository since the last release. They are made available in the [Changelog](#changelog), in Git Tags and through the [forge](#forge)-native Releases.

Learn more in the [Release Notes customization](../guides/release-notes.md) guide.

### Semantic Versioning (SemVer)

[Semantic Versioning](https://semver.org/) is a specification for version numbers. It is the only supported versioning schema in
`releaser-pleaser`. Follow the link to learn more.

### Updater

Updaters can update or create files that will be included in [Release Pull Request](#release-pull-request). Examples of Updaters are
`changelog` for `CHANGELOG.md`, `generic` that can update arbitrary files,
`packagejson` that knows how to update Node.JS `package.json` files and
`helmchart` for Helm's `Chart.yaml` file.