# {{.Name}}

<dl>
  <dt>Go module path:</dt>
  <dd><code>{{.ModulePath}}</code></dd>

  <dt>Repository:</dt>
  <dd><a href="{{.RepoURL}}">{{.RepoURL}}</a>{{if .SubdirURL}}, subdir <a href="{{.SubdirURL}}">{{.SubdirName}}</a>{{end}}</dd>
</dl>

{{.Description}}

::: section small
(This page is automatically generated to make <code>go install {{.ModulePath}}</code> work.)
:::
