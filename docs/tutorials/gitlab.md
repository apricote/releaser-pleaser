# Getting started on GitLab

In this tutorial you will learn how to set up `releaser-pleaser` in your GitLab project with GitLab CI.

> In `releaser-pleaser` documentation we mostly use "Pull Request" (GitHub wording) instead of "Merge Request" (GitLab wording). The GitLab-specific pages are an exception and use "Merge Request".

## 1. Project Settings

### 1.1. Merge Requests

`releaser-pleaser` requires _Fast-forward merges_ and _squashing_. With other merge options it can not reliably find the right merge request for every commit on `main`.

Open your project settings to page _Merge Requests_:

> `https://gitlab.com/YOUR-PATH/YOUR-PROJECT/-/settings/merge_requests`

In the "Merge method" section select "Fast-forward merge":

![Screenshot of the required merge method settings](./gitlab-settings-merge-method.png)

In the "Squash commits when merging" section select "Require":

![Screenshot of the required squash settings](./gitlab-settings-squash.png)

## 2. API Access Token

`releaser-pleaser` uses the GitLab API to create the [release merge request](../explanation/release-pr.md) and subsequent releases for you. The default `GITLAB_TOKEN` available in CI jobs does not have enough permissions for this, so we need to create an Access Token and make it available in a CI variable.

## 2.1. Create Project Access Token

Open your project settings to page _Access tokens_:

> `https://gitlab.com/YOUR-PATH/YOUR-PROJECT/-/settings/access_tokens`

Create a token with these settings:

- **Name**: `releaser-pleaser`
- **Role**: `Maintainer`
- **Scopes**: `api`, `read_repository`, `write_repository`

Copy the created token for the next step.

![Screenshot of the access token settings](./gitlab-access-token.png)

## 2.2. Save token in CI variable

Open your project settings to page _CI/CD_:

> `https://gitlab.com/YOUR-PATH/YOUR-PROJECT/-/settings/ci_cd`

In the section "Variables" click on the "Add variable" button to open the form for a new variable. Use these settings to create the new variable:

- **Type**: Variable
- **Visibility**: Masked
- **Flags**: Uncheck "Protect variable" if your `main` branch is not protected
- **Key**: `RELEASER_PLEASER_TOKEN`
- **Value**: The project access token from the previous step

## 3. GitLab CI/CD

`releaser-pleaser` is published as a [GitLab CI/CD Component](https://docs.gitlab.com/ee/ci/components/): https://gitlab.com/explore/catalog/apricote/releaser-pleaser

Create or open your `.gitlab-ci.yml` and add the following include to your configuration:

```yaml
stages: [build]

include:
  - component: $CI_SERVER_FQDN/apricote/releaser-pleaser/run@v0.4.0-beta.1
    inputs:
      token: $RELEASER_PLEASER_TOKEN
```

> You can set the `stage` input if you want to run `releaser-pleaser` during a different stage.

## 4. Release Merge Request

Once the `releaser-pleaser` job runs for the first time, you can check the logs to see what it did.
If you have releasable commits since the last tag, `releaser-pleaser` opens a release merge request for the proposed release.

Once you merge this merge request, `releaser-pleaser` automatically creates a Git tag and GitLab Release with the proposed version and changelog.

## Related Documentation

- **Explanation**
  - [Release Pull Request](../explanation/release-pr.md)
- **Reference**
  - [GitLab CI/CD Component](../reference/gitlab-cicd-component.md)
