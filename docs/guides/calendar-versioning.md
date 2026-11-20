# Calendar Versioning

By default, `releaser-pleaser` creates [Semantic Versions](../reference/glossary.md#semantic-versioning-semver). It can create [Calendar Versions](../reference/glossary.md#calendar-versioning-calver) instead:

| Option          | Values                              | Default         |
| --------------- | ----------------------------------- | --------------- |
| `versioning`    | `semver`, `calver`                  | `semver`        |
| `calver-format` | see [CalVer format](#calver-format) | `YYYY.WW.MICRO` |

Both are available as inputs of the [GitHub Action](../reference/github-action.md) and the [GitLab CI/CD Component](../reference/gitlab-cicd-component.md), and as `--versioning` and `--calver-format` flags of `rp run`.

## Example

GitHub Actions:

```yaml
- uses: apricote/releaser-pleaser@v0.9.0
  with:
    versioning: calver
    calver-format: YYYY.WW.MICRO
```

GitLab CI/CD:

```yaml
include:
  - component: $CI_SERVER_FQDN/apricote/releaser-pleaser/run@v0.9.0
    inputs:
      token: $RELEASER_PLEASER_TOKEN
      versioning: calver
      calver-format: YYYY.WW.MICRO
```

## Versions

With `versioning: calver` the version is built from the current date and a counter, following the [CalVer](https://calver.org/) conventions. With the default format `YYYY.WW.MICRO`:

- The first release in ISO week 47 of 2026 is `2026.47.0`.
- Further releases in the same week increase the counter: `2026.47.1`, `2026.47.2`, …
- The first release in the next week is `2026.48.0`.
- Weeks are not zero-padded: the first release in ISO week 7 is `2026.7.0`. Use `YYYY.0W.MICRO` for `2026.07.0`.

CalVer versions never have a `v` prefix. Tags that do not match the format (for example older SemVer tags like `v1.2.3`) are ignored. If you switch an existing repository to CalVer, the first CalVer release contains all commits since the repository was created.

## CalVer format

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

Examples: `YYYY.WW.MICRO`, `YYYY.0W.MICRO`, `YYYY.0M.MICRO`, `YY.MM.DD.MICRO`.

All dates are in UTC. If the format contains a week token, the year tokens use the ISO week-numbering year. For example, 2027-01-01 belongs to week 53 of 2026, so the version is `2026.53.0` and not `2027.53.0`.

## Pre-releases

The [pre-release](pre-releases.md) labels also work with CalVer. They add a suffix to the version: `2026.47.0-rc.0`, `2026.47.0-rc.1`, … The stable release afterwards is `2026.47.0`. If the date changes, the counter starts over at `rc.0` for the new date.

## Caveats

- **The version is calculated when the release pull request is updated, not when it is merged.** `releaser-pleaser` runs and updates the pull request on every push to your branch. If the release pull request is opened in week 47 and merged in week 48 without any other push in between, the release is still called `2026.47.x`. If this matters to you, also run `releaser-pleaser` on a schedule, e.g. daily, so that the version in the pull request is updated.
- **`packagejson` updater**: npm requires valid SemVer versions. Zero-padded tokens (`0W`, `0M`, `0D`, `0Y`) produce versions like `2026.07.0`, which npm rejects. Use the unpadded tokens if you use the `packagejson` updater.

## Related Documentation

- **Guides**
  - [Pre-releases](pre-releases.md)
  - [Customizing Release Notes](release-notes.md)
- **Reference**
  - [Versioning](../reference/versioning.md)
  - [GitHub Action](../reference/github-action.md)
  - [GitLab CI/CD Component](../reference/gitlab-cicd-component.md)
