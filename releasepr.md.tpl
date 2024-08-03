{{ .Changelog }}

---

## releaser-pleaser Instructions
{{ if .Overrides }}
{{- .Overrides -}}
{{- else }}
<!-- section-start overrides -->
> If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

### Prefix

```rp-prefix
```

### Suffix

```rp-suffix
```

<!-- section-end overrides -->

{{ end }}
#### PR by [releaser-pleaser](https://github.com/apricote/releaser-pleaser)
