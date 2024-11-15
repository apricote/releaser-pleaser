# Changelog

## [v0.5.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.5.0)

### Features

- **gitlab**: make job dependencies configurable and run immediately (#101)
- **github**: mark pre-releases correctly (#110)

### Bug Fixes

- use commits with slightly invalid messages in release notes (#105)
- create CHANGELOG.md if it does not exist (#108)

## [v0.4.2](https://github.com/apricote/releaser-pleaser/releases/tag/v0.4.2)

### Bug Fixes

- **action**: container image reference used wrong syntax (#96)

## [v0.4.1](https://github.com/apricote/releaser-pleaser/releases/tag/v0.4.1)

### Bug Fixes

- **gitlab**: release not created when release pr was squashed (#86)

## [v0.4.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.4.0)

### ✨ Highlights

#### GitLab Support

You can now use `releaser-pleaser` with projects hosted on GitLab.com and self-managed GitLab installations. Check out the new [tutorial](https://apricote.github.io/releaser-pleaser/tutorials/gitlab.html) to get started.

### Features

- add support for GitLab repositories (#49)
- add shell to container image (#59)
- **gitlab**: add CI/CD component (#55)
- **changelog**: omit version heading in forge release notes
- **gitlab**: support self-managed instances (#75)

### Bug Fixes

- **parser**: continue on unparsable commit message (#48)
- **cli**: command name in help output (#52)
- **parser**: invalid handling of empty lines (#53)
- multiple extra-files are not evaluated properly (#61)

## [v0.4.0-beta.1](https://github.com/apricote/releaser-pleaser/releases/tag/v0.4.0-beta.1)

### Features

- add shell to container image (#59)
- **gitlab**: add CI/CD component (#55)

### Bug Fixes

- multiple extra-files are not evaluated properly (#61)

## [v0.4.0-beta.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.4.0-beta.0)

### Features

- add support for GitLab repositories (#49)

### Bug Fixes

- **parser**: continue on unparsable commit message (#48)
- **cli**: command name in help output (#52)
- **parser**: invalid handling of empty lines (#53)

## [v0.3.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.3.0)

### :sparkles: Highlights

#### Cleaner pre-release Release Notes

From now on if you create multiple pre-releases in a row, the release notes will only include changes since the last pre-release. Once you decide to create a stable release, the release notes will be in comparison to the last stable release.

#### Edit pull request after merging.

You can now edit the message for a pull request after merging by adding a \```rp-commits code block into the pull request body. Learn more in the [Release Notes Guide](https://apricote.github.io/releaser-pleaser/guides/release-notes.html#editing-the-release-notes).

### Features

- less repetitive entries for prerelease changelogs #37
- format markdown in changelog entry (#41)
- edit commit message after merging through PR (#43)
- **cli**: show release PR url in log messages (#44)

## [v0.2.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.2.0)

### Features

- update version references in any files (#14)
- **cli**: add --version flag (#29)

### Bug Fixes

- **ci**: building release image fails (#21)
- **ci**: ko pipeline permissions (#23)
- **action**: invalid quoting for extra-files arg (#25)

## [v0.2.0-beta.2](https://github.com/apricote/releaser-pleaser/releases/tag/v0.2.0-beta.2)

### Features

- update version references in any files (#14)

### Bug Fixes

- **ci**: building release image fails (#21)
- **ci**: ko pipeline permissions (#23)

## [v0.2.0-beta.1](https://github.com/apricote/releaser-pleaser/releases/tag/v0.2.0-beta.1)

### Features

- update version references in any files (#14)

### Bug Fixes

- **ci**: building release image fails (#21)

## [v0.2.0-beta.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.2.0-beta.0)

### Features

- update version references in any files (#14)

## [v0.1.0](https://github.com/apricote/releaser-pleaser/releases/tag/v0.1.0)

### This is the first release ever, so it also includes a lot of other functionality.

### Features

- add github action (#1)
