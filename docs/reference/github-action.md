# GitHub Action

## Reference

The action is available as `apricote/releaser-pleaser` on GitHub.com.

## Versions

The `apricote/releaser-pleaser` action is released together with `releaser-pleaser` and they share the version number.

The action does not support floating tags (e.g. `v1`) right now ([#31](https://github.com/apricote/releaser-pleaser/issues/31)). You have to use the full version or commit SHA instead: `apricote/releaser-pleaser@v0.2.0`.

## Inputs

The following inputs are supported by the `apricote/releaser-pleaser` GitHub Action.

| Input         | Description                                            |         Default |                                                              Example |
| ------------- | :----------------------------------------------------- | --------------: | -------------------------------------------------------------------: |
| `branch`      | This branch is used as the target for releases.        |          `main` |                                                             `master` |
| `token`       | GitHub token for creating and updating release PRs     | `$GITHUB_TOKEN` |                                `${{secrets.RELEASER_PLEASER_TOKEN}}` |
| `extra-files` | List of files that are scanned for version references. |            `""` | <pre><code>version/version.go<br>deploy/deployment.yaml</code></pre> |

## Outputs

The action does not define any outputs.
