# Updaters

There are different updater for different purposes available.

They each have a name and may be enabled by default. You can configure which updaters are used through the
`updaters` input on GitHub Actions and GitLab CI/CD. This is a comma-delimited list of updaters that should be enabled, for updaters that are enabled by default you can remove them by adding a minus before its name:

```
updaters: -generic,packagejson
```

## Changelog

- **Name**: `changelog`
- **Default**: enabled

This updater creates the `CHANGELOG.md` file and adds new release notes to it.

## Generic Updater

- **Name**: `generic`
- **Default**: enabled

This updater can update any file and only needs a marker on the line. It is enabled by default.

Learn more about this updater in ["Updating arbitrary files"](../guides/updating-arbitrary-files.md).

## Node.js `package.json` Updater

- **Name**: `packagejson`
- **Default**: disabled

This updater can update the `version` field in Node.js `package.json` files. The updater is disabled by default.

## Helm's `Chart.yaml` Updater

- **Name**: `helmchart`
- **Default**: disabled

This updater can update the `version` field in Helm's `Chart.yaml` files. The updater is disabled by default.
