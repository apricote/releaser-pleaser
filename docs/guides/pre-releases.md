# Pre-releases

Pre-releases are a concept of [SemVer](#semantic-versioning-semver). They follow the normal versioning schema but use a suffix out of `-alpha.X`, `-beta.X` and `-rc.X`.

Pre-releases are not considered "stable" and are usually not recommended for most users.

## Creating a pre-release

If you want to create a pre-release, you can set **one** of the following labels on the release pull request:

- `rp-next-version::alpha`
- `rp-next-version::beta`
- `rp-next-version::rc`

This will cause `releaser-pleaser` to run, and it will change the release pull request to a matching version according to the type of pre-release.

## Versioning

For pre-releases, `releaser-pleaser` analyzes the commits made since the **last stable release**. The version bump from this is then applied to the last stable release and the pre-release info is added to the version number. If a previous pre-release of the matching type exists, the "pre-release counter" at the end of the version is increased by one.

An examples:

- The last stable version was `v1.0.0`
- Since then a `feat` commit was merged, this causes a bump of the minor version: `v1.1.0`
- The release pull request has the label `rp-next-version::beta`. This changes the suggested version to `v1.1.0-beta.0`

If there was already a `v1.1.0-beta.0`, then the suggested version would be `v1.1.0-beta.1`.

Changing the pre-release type (for example from `beta` to `rc`), resets the counter. `v1.1.0-beta.1` would be followed by `v1.1.0-rc.0`.

## Stable Release

`releaser-pleaser` ignores pre-releases when looking for releasable commits. This means that right after creating a new pre-release, `releaser-pleaser` again detects releasable commits and opens a new release pull request for the stable version.

## Related Documentation

- **Reference**
  - [Pull Request Options](../reference/pr-options.md)
