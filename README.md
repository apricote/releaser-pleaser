# releaser-pleaser

`releaser-pleaser` is a tool designed to automate versioning and changelog management for your projects. Building on the concepts of [
`release-please`](https://github.com/googleapis/release-please), it streamlines the release process through GitHub Actions or GitLab CI.

## Features

- **Automated Pull Requests**: Opens a PR when releasable changes are detected.
- **Smart Versioning**: Suggests new versions based on conventional commits and semantic versioning.
- **Version Reference Updates**: Automatically updates committed version references in the PR.
- **Changelog Generation**: Creates new changelog entries based on commits.
- **Automated Releases**: Upon PR merge, creates tags and GitHub/GitLab Releases with appropriate messages.
- **Version Overrides**: Allows manual override of the suggested version.
- **Prerelease Support**: Offers options to create alpha, beta, or release candidate versions.

`releaser-pleaser` simplifies release management, allowing maintainers to focus on development while ensuring consistent and well-documented releases.

## Status

This project is still under active development. You can not reasonably use it right now and not all features advertised above work. Keep your eyes open for any releases.

## Relation to `release-please`

After using
`release-please` for 1.5 years, I've found it to be the best tool for low-effort releases currently available. While I appreciate many of its features, I identified several additional capabilities that would significantly enhance my workflow. Although it might be possible to incorporate these features into
`release-please`, I decided to channel my efforts into creating a new tool that specifically addresses my needs.

Key differences in `releaser-pleaser` include:

- Support for multiple forges (both GitHub and GitLab)
- Better support for pre-releases

One notable limitation of
`release-please` is its deep integration with the GitHub API, making the addition of support for other platforms (like GitLab) a substantial undertaking.
`releaser-pleaser` aims to overcome this limitation by design, offering a more versatile solution for automated release management across different platforms and project requirements.

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0).
