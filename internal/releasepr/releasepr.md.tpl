<!-- section-start changelog -->
{{ .Changelog }}
<!-- section-end changelog -->

---

<details>
  <summary><h4>PR by <a href="https://github.com/apricote/releaser-pleaser">releaser-pleaser</a> 🤖</h4></summary>

If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

## Release Notes

### Prefix / Start

This will be added to the start of the release notes.

```rp-prefix
{{- if .Overrides.Prefix }}
{{ .Overrides.Prefix }}{{ end }}
```

### Suffix / End

This will be added to the end of the release notes.

```rp-suffix
{{- if .Overrides.Suffix }}
{{ .Overrides.Suffix }}{{ end }}
```

</details>
