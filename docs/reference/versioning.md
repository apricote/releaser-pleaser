# Versioning

`releaser-pleaser` supports different versioning schemes. The scheme decides the version of the next release, and which existing tags are considered releases.

You can configure the versioning scheme through the `versioning` input on GitHub Actions and GitLab CI/CD, or the `--versioning` flag of `rp run`:

```
versioning: calver
calver-format: YYYY.WW.MICRO
```

Tags that are not valid versions of the configured scheme are ignored.

## Semantic Versioning

- **Name**: `semver`
- **Default**: yes
- **Example**: `v1.2.3`, `v1.3.0-rc.0`

The next version is calculated from the [Conventional Commits](glossary.md#conventional-commits) since the last release:

| Commits                         | Bump  | Example             |
| ------------------------------- | ----- | ------------------- |
| Breaking change (e.g. `feat!:`) | Major | `v1.2.3` → `v2.0.0` |
| `feat`                          | Minor | `v1.2.3` → `v1.3.0` |
| `fix`                           | Patch | `v1.2.3` → `v1.2.4` |

The [`scoped` commit parser](commit-parsers.md#scoped-commits) can not be used with SemVer, because its commits do not tell which part of the version to bump.

Versions always have a `v` prefix. If there is no previous release, the version is bumped from `v0.0.0`.

Learn more about the scheme on [semver.org](https://semver.org/).

## Calendar Versioning

- **Name**: `calver`
- **Default**: no
- **Example**: `2026.47.0`, `2026.47.1-rc.0`

The version is built from the current date (in UTC) and a counter (`MICRO`) according to the `calver-format` option. The counter starts at `0` for each date and increases with every release for the same date. Versions never have a `v` prefix.

The type of the commits does not affect the version, every releasable commit leads to a new release.

Learn more in the [Calendar Versioning](../guides/calendar-versioning.md) guide.

### Format

- **Option**: `calver-format` (input), `--calver-format` (flag of `rp run`)
- **Default**: `YYYY.WW.MICRO`

The format consists of tokens separated by `.`:

| Token   | Description                        | Example (2026-02-11) |
| ------- | ---------------------------------- | -------------------: |
| `YYYY`  | Full year                          |               `2026` |
| `YY`    | Short year, since 2010             |                 `26` |
| `0Y`    | Zero-padded short year             |                 `26` |
| `MM`    | Month                              |                  `2` |
| `0M`    | Zero-padded month                  |                 `02` |
| `WW`    | ISO week                           |                  `7` |
| `0W`    | Zero-padded ISO week               |                 `07` |
| `DD`    | Day                                |                 `11` |
| `0D`    | Zero-padded day                    |                 `11` |
| `MICRO` | Counter, starts at 0 for each date |                  `0` |

Rules:

- The format must end with `MICRO`, and `MICRO` can only appear once.
- Each part of the date (year, month, week, day) can only appear once.
- Week tokens can not be combined with month or day tokens.
- If the format contains a week token, the year tokens use the ISO week-numbering year.

## Pre-releases

Both schemes support [pre-releases](../guides/pre-releases.md) through the `rp-next-version::*` [labels](pr-options.md#release-type) on the release pull request. They add a suffix to the version:

| Label                     | SemVer           | CalVer              |
| ------------------------- | ---------------- | ------------------- |
| `rp-next-version::alpha`  | `v1.3.0-alpha.0` | `2026.47.0-alpha.0` |
| `rp-next-version::beta`   | `v1.3.0-beta.0`  | `2026.47.0-beta.0`  |
| `rp-next-version::rc`     | `v1.3.0-rc.0`    | `2026.47.0-rc.0`    |
| `rp-next-version::normal` | `v1.3.0`         | `2026.47.0`         |

Further pre-releases of the same type increase the number in the suffix, e.g. `-rc.1`.
