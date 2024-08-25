# Customizing Release Notes

You can customize the generated Release Notes in two ways:

## For a single commit / pull request

This feature is still being worked on. Check out [#5](https://github.com/apricote/releaser-pleaser/issues/5) for the current status.

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
