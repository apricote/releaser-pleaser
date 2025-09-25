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

**Examples**:

    ~~~~rp-prefix
    #### Awesome new feature!

    This text is at the start of the release notes.
    ~~~~

    ~~~~rp-suffix
    #### Version Compatibility

    And this at the end.
    ~~~~

### Status

**Labels**:

- `rp-release::pending`
- `rp-release::tagged`

These labels are automatically added by `releaser-pleaser` to release pull requests. They are used to track if the corresponding release was already created.

Users should not set these labels themselves.

## Other Pull Requests

Not created by `releaser-pleaser`.

### Release Notes

**Code Blocks**:

- `rp-commits`

If specified, `releaser-pleaser` will consider each line in the code block as a commit message and add all of them to the Release Notes. Learn more in the [Release Notes](../guides/release-notes.md) guide.

The types of commits (`feat`, `fix`, ...) are also considered for the next version.

**Examples**:

    ```rp-commits
    feat(api): add movie endpoints
    fix(db): invalid schema for actor model
    ```
