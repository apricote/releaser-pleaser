# Customizing Release Notes

You can customize the generated Release Notes in two ways:

## For a single commit / pull request

### Editing the Release Notes

After merging a non-release pull request, you can still modify how it appears in the Release Notes.

To do this, add a code block named `rp-commits` in the pull request description. When this block is present, `releaser-pleaser` will use its content for generating Release Notes instead of the commit message. If the code block contains multiple lines, each line will be treated as if it came from separate pull requests. This is useful for pull requests that introduce multiple features or fix several bugs.

You can update the description at any time after merging the pull request but before merging the release pull request. `releaser-pleaser` will then re-run and update the suggested Release Notes accordingly.

>     ```rp-commits
>     feat(api): add movie endpoints
>     feat(api): add cinema endpoints
>     fix(db): invalid schema for actor model
>     ```

Using GitHub as an example, the pull request you are trying to change the Release Notes for should look like this:

![Screenshot of a pull request page on GitHub. Currently editing the description of the pull request and adding the rp-commits snippet from above.](release-notes-rp-commits.png)

In turn, `releaser-pleaser` updates the release pull request like this:

![Screenshot of a release pull request on GitHub. It shows the release notes with the three commits from the rp-commits example.](release-notes-rp-commits-release-pr.png)

### Removing the pull request from the Release Notes

If you add an empty code block, the pull request will be removed from the Release Notes.

>     ```rp-commits
>     ```

## For the release

It is possible to add custom **prefix** and **suffix** Markdown-formatted text to the Release Notes.

The release pull request description has text fields where maintainers can add the prefix and suffix. To see these fields, toggle the collapsible section in the description:

![Screenshot of the collapsed section](./release-notes-collapsible.png)

When you edit the description, make sure to put your desired content into the code blocks named `rp-prefix` and `rp-suffix`. Only the content of these blocks is considered.

>     ```rp-prefix
>     ### Prefix
>
>     This will be shown as the Prefix.
>     ```
>
>     ```rp-suffix
>     ### Suffix
>
>     This will be shown as the Suffix.
>     ```

To match the style of the auto-generated release notes, you should start any headings at level 3 (`### Title`).

Once the description was updated `releaser-pleaser` automatically runs again and adds the prefix and suffix to the Release Notes and to the committed Changelog:

```markdown
## v1.1.0

### Prefix

This will be shown as the Prefix.

### Features

- Added cool new thing (#1)

### Suffix

This will be shown as the Suffix.
```

## Related Documentation

- **Reference**
  - [Pull Request Options](../reference/pr-options.md)
