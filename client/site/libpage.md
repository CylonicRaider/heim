# {{.Name}}

Go module path:
: `{{.ModulePath}}`

VCS:
: `{{.VCS}}`

Repository:
: <a href="{{.RepoURL}}">{{.RepoURL}}</a>{{if .SubdirName}}, subdir <a href="{{.SubdirURL}}">{{.SubdirName}}</a>{{end}}

{{.Description}}

::: section small
(This page is automatically generated to make `go install {{.ModulePath}}` work.)
:::
