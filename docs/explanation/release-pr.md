# Release Pull Request

A _release pull request_ is opened by `releaser-pleaser` when it detects that there are _releasable changes_.
The pull request contains an _auto-generated Changelog_ and a _suggested next version_.
Once someone merges this pull request, `releaser-pleaser` will create a matching Git Tag and Release on GitHub/GitLab.

Maintainers can fill various fields in the pull request description and through labels to change the proposed release. Some examples of this are: _Changelog Prefix & Suffix text_ and _requesting a pre-release_ (`alpha`, `beta`, `rc`) version.

The pull request is automatically updated by `releaser-pleaser` every time it runs.

### Example Screenshot

![Screenshot of an example Release Pull Request on GitHub](./release-pr.png)

## Related Documentation

- **Guide**
  - [Pre-releases](../guides/pre-releases.md)
  - [Release Notes](../guides/release-notes.md)
- **Reference**
  - [Pull Request Options](../reference/pr-options.md)
