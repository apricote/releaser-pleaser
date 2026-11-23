# Signing release commits

For every release pull request, `releaser-pleaser` creates one commit, the release commit, and pushes it to the `releaser-pleaser--branches--<branch>` branch. If you give it an OpenPGP private key, it signs that commit.

You need this when your forge rejects unsigned commits, for example through a GitLab push rule ("Reject unsigned commits") or a GitHub ruleset ("Require signed commits"). It also lets the forge show the release commit as verified.

The release commit is the only commit `releaser-pleaser` creates. When you merge the release pull request, the forge creates the commit on your target branch, so the forge decides how that commit is signed.

## 1. Create a key

`releaser-pleaser` commits as the user that owns its access token, with the name and email the forge API returns for that user. Create the key for the same email, so the forge can match the signature to the user:

```shell
gpg --quick-generate-key "releaser-pleaser <EMAIL>" ed25519 sign never
gpg --armor --export-secret-keys EMAIL > signing-key.asc
gpg --armor --export EMAIL > signing-key.pub.asc
```

A passphrase is optional. If the key has one, pass it as well (see below).

## 2. Add the public key to the user

Add `signing-key.pub.asc` to the GPG keys of the user that owns the token. Forges show a commit as verified only when the signing key belongs to the committer and its email matches one of the user's verified email addresses.

## 3. Pass the private key to releaser-pleaser

`releaser-pleaser` reads the key from these environment variables:

| Variable                                  | Description                                                                         |
| ----------------------------------------- | ----------------------------------------------------------------------------------- |
| `RELEASER_PLEASER_SIGNING_KEY`            | The ASCII-armored private key                                                       |
| `RELEASER_PLEASER_SIGNING_KEY_FILE`       | Path to a file holding the ASCII-armored private key. Set this or the one above.    |
| `RELEASER_PLEASER_SIGNING_KEY_PASSPHRASE` | Passphrase of the key, if it is encrypted                                           |

When none of them are set, the release commit is unsigned. When a key is set, the job log shows `signing release commits` with the key ID.

### GitHub Actions

Store the key and its passphrase as repository secrets and pass them to the action, together with the token of the user that owns the key:

```yaml
- uses: apricote/releaser-pleaser@v0.9.0
  with:
    token: ${{ secrets.RELEASER_PLEASER_TOKEN }}
    signing-key: ${{ secrets.RELEASER_PLEASER_SIGNING_KEY }}
    signing-key-passphrase: ${{ secrets.RELEASER_PLEASER_SIGNING_KEY_PASSPHRASE }}
```

### GitLab CI/CD

GitLab can only mask single-line values, so store the key as a variable of type **File** named `RELEASER_PLEASER_SIGNING_KEY_FILE`. GitLab writes the key to a file in the job and sets the variable to its path. Store the passphrase as a masked variable named `RELEASER_PLEASER_SIGNING_KEY_PASSPHRASE`. Mark both as protected, so only pipelines on protected branches receive them.

The component job reads both variables from its environment; no inputs are needed.

## Related Documentation

- **Reference**
  - [GitHub Action](../reference/github-action.md)
  - [GitLab CI/CD Component](../reference/gitlab-cicd-component.md)
