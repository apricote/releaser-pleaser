# GitLab CI/CD Component

## Reference

The CI/CD component is available as `$CI_SERVER_FQDN/apricote/releaser-pleaser/run` on gitlab.com.

It is being distributed through the CI/CD Catalog: [apricote/releaser-pleaser](https://gitlab.com/explore/catalog/apricote/releaser-pleaser).

## Versions

The `apricote/releaser-pleaser` action is released together with `releaser-pleaser` and they share the version number.

The component does not support floating tags (e.g.
`v1`) right now ([#31](https://github.com/apricote/releaser-pleaser/issues/31)). You have to use the full version or commit SHA instead:
`apricote/releaser-pleaser@v0.4.0`.

## Inputs

The following inputs are supported by the component.

| Input                  | Description                                               | Default |                                                              Example |
| ---------------------- | :-------------------------------------------------------- | ------: | -------------------------------------------------------------------: |
| `branch`               | This branch is used as the target for releases.           |  `main` |                                                             `master` |
| `token` (**required**) | GitLab access token for creating and updating release PRs |         |                                            `$RELEASER_PLEASER_TOKEN` |
| `extra-files`          | List of files that are scanned for version references.    |    `""` | <pre><code>version/version.go<br>deploy/deployment.yaml</code></pre> |
| `stage`                | Stage the job runs in. Must exists.                       | `build` |                                                               `test` |
| `needs`                | Other jobs the releaser-pleaser job depends on.           |    `[]` |              <pre><code>- validate-foo<br>- prepare-bar</code></pre> |
