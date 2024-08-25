# Pull Request Options

The proposed releases can by influenced by changing the description and labels of either the release pull request or the normal pull requests created by other developers. This document lists the available options for both types of pull requests.

## Release Pull Request

Created by `releaser-pleaser`.

### Release Type

**Labels**:

- `rp-next-version::alpha`
- `rp-next-version::beta`
- `rp-next-version::rc`
- `rp-next-version::normal`

Adding one of these labels will change the type of the next release to the one indicated in the label. This is used to create [pre-releases](../guides/pre-releases.md).

Adding more than one of these labels is not allowed and the behaviour if multiple labels are added is undefined.

### Release Notes

**Code Blocks**:

- `rp-prefix`
- `rp-suffix`

Any text in code blocks with these languages is being added to the start or end of the Release Notes and Changelog. Learn more in the [Release Notes](../guides/release-notes.md) guide.

### Status

**Labels**:

- `rp-release::pending`
- `rp-release::tagged`

These labels are automatically added by `releaser-pleaser` to release pull requests. They are used to track if the corresponding release was already created.

Users should not set these labels themselves.

## Other Pull Requests

Not created by `releaser-pleaser`.

Normal pull requests do not support any options right now.
