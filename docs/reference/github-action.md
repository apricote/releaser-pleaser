# GitHub Action

## Reference

The action is available as `apricote/releaser-pleaser` on GitHub.com.

## Versions

The `apricote/releaser-pleaser` action is released together with `releaser-pleaser` and they share the version number.

The action does not support floating tags (e.g.
`v1`) right now ([#31](https://github.com/apricote/releaser-pleaser/issues/31)). You have to use the full version or commit SHA instead:
`apricote/releaser-pleaser@v0.2.0`.

## Inputs

The following inputs are supported by the `apricote/releaser-pleaser` GitHub Action.

| Input                    | Description                                                                                                                                                                            |                Default |                                                              Example |
| ------------------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------: | -------------------------------------------------------------------: |
| `branch`                 | This branch is used as the target for releases.                                                                                                                                        |                 `main` |                                                             `master` |
| `token`                  | GitHub token for creating and updating release PRs                                                                                                                                     |        `$GITHUB_TOKEN` |                                `${{secrets.RELEASER_PLEASER_TOKEN}}` |
| `forge`                  | Forge this action is run against                                                                                                                                                       |               `github` |                                                            `forgejo` |
| `extra-files`            | List of files that are scanned for version references by the generic updater.                                                                                                          |                   `""` | <pre><code>version/version.go<br>deploy/deployment.yaml</code></pre> |
| `updaters`               | List of updaters that are run. Default updaters can be removed by specifying them as -name. Multiple updaters should be concatenated with a comma. Default Updaters: changelog,generic |                   `""` |                                               `-generic,packagejson` |
| `commit-parser`          | Format of the commit messages. One of `conventional-commits`, `scoped`. See [Commit Parsers](commit-parsers.md).                                                                       | `conventional-commits` |                                                             `scoped` |
| `versioning`             | Versioning scheme. One of `semver`, `calver`. See [Calendar Versioning](../guides/calendar-versioning.md).                                                                             |               `semver` |                                                             `calver` |
| `calver-format`          | Format of CalVer versions, only used with `versioning: calver`.                                                                                                                        |        `YYYY.WW.MICRO` |                                                      `YYYY.0M.MICRO` |
| `api-url`                | API URL of the forge this action is run against.                                                                                                                                       |                   `""` |                                        `https://forgejo.example.com` |
| `owner`                  | Owner of the repository. Only required for Forgejo Actions.                                                                                                                            |                   `""` |                                                           `apricote` |
| `repo`                   | Name of the repository. Only required for Forgejo Actions.                                                                                                                             |                   `""` |                                                   `releaser-pleaser` |
| `signing-key`            | ASCII-armored OpenPGP private key used to sign the release commit. See [Signing release commits](../guides/signed-commits.md).                                                         |                   `""` |                          `${{secrets.RELEASER_PLEASER_SIGNING_KEY}}` |
| `signing-key-passphrase` | Passphrase of the signing key, if it is encrypted.                                                                                                                                     |                   `""` |               `${{secrets.RELEASER_PLEASER_SIGNING_KEY_PASSPHRASE}}` |

## Outputs

The action does not define any outputs.
