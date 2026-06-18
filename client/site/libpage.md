# {{.Name}}

<dl>
  <dt>Go module path:</dt>
  <dd><code>{{.ModulePath}}</code></dd>

  <dt>VCS:<dt>
  <dd><code>{{.VCS}}</code></dd>

  <dt>Repository:</dt>
  <dd><a href="{{.RepoURL}}">{{.RepoURL}}</a>{{if .SubdirName}}, subdir <a href="{{.SubdirURL}}">{{.SubdirName}}</a>{{end}}</dd>
</dl>

{{.Description}}

::: section small
(This page is automatically generated to make <code>go install {{.ModulePath}}</code> work.)
:::
