# Workflow Permissions on GitHub

## Default GitHub token permissions

The [GitHub](../tutorials/github.md) tutorial uses the builtin `GITHUB_TOKEN` for the action to get access to the repository. It uses the following permissions on the token:

```yaml
jobs:
  releaser-pleaser:
    permissions:
      # - list commits
      # - push commits for the release pull request
      # - push new releases & tags
      contents: write

      # - read pull requests for Changelog
      # - read and write release pull request
      # - create labels on the repository
      pull-requests: write
```

These permissions are sufficient for simple operations. But fail if you want to run another workflow on `push: tag`.

## Workflows on Tag Push

When using the automatic `GITHUB_TOKEN` to create tags, GitHub does not create new workflow runs that are supposed to be created. This is done to prevent the user from "accidentally creating recursive workflow runs". You can read more about this behaviour in the [GitHub Actions docs](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication#using-the-github_token-in-a-workflow).

Workflows that have a trigger on pushed tags are often used to build artifacts for the release, like binaries or container images.

```yaml
on:
  push:
    tags:
      - "v*.*.*"
```

To circumvent this restriction, you can create a personal access token and instruct `releaser-pleaser` to use that instead to talk to GitHub.

### 1. Create Personal Access Token

On your account settings, navigate to the [Personal access tokens](https://github.com/settings/tokens?type=beta) section.

You can either use a "Fine-grained" or "Classic" token for this. Fine-grained tokens can be restricted to specific actions and repositories and are more secure because of this. On the other hand they have a mandatory expiration of 1 year maximum. Classic tokens have unrestricted access to your account, but do not expire.

Copy the token for the next step.

#### Fine-grained token

When you create a fine-grained token, restrict the access to the repository where you are using `releaser-pleaser`.

In the **repository permissions** you need to give **read and write** access for **Contents** and **Pull requests**. All other permissions can be set to **No access** (default).

No **account permissions** are required and you can set all to **No access** (default).

### 2. Repository Secret

Next you need to add the personal access token as a repository secret.

Open the repository settings to **Secrets and variables > Actions**:

> `https://github.com/YOUR-NAME/YOUR-REPO/settings/secrets/actions`

Click on **New repository secret** and add the personal access token to a secret named `RELEASER_PLEASER_TOKEN`.

### 3. Update Workflow

Update the workflow file (`.github/workflows/releaser-pleaser.yaml`) to pass the new secret to the `releaser-pleaser` action. You can also remove the permissions of the job, as they are now unused.

```diff
 jobs:
   releaser-pleaser:
     runs-on: ubuntu-latest # The action uses docker containers
-    permissions:
-      contents: write
-      pull-requests: write
     steps:
       - name: releaser-pleaser
         uses: apricote/releaser-pleaser@v0.2.0
+        with:
+          token: ${{ secrets.RELEASER_PLEASER_TOKEN }}
```

The next release created by releaser-pleaser will now create the follow-up workflows as expected.
