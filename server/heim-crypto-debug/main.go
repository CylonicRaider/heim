package main

import (
	"sort"
	"strings"
)

type params struct {
	Name     string `usage:"User name."`
	Password string `usage:"Password (as plain text)."`
	Payload  []byte `usage:"Binary payload." cli:",arg,required"`
	Count    int    `usage:"Repetition count." cli:",arg"`
}

func (p *params) Run(env CLIEnv) {
	if p.Name == "" {
		p.Name = *env.ReadLine("Username: ")
	}
	if p.Password == "" {
		p.Password = *env.ReadPass("Password: ")
	}

	if p.Password == "hunter2" {
		env.Println("Hello " + p.Name + "!")
		for i := 0; i < p.Count; i++ {
			if i != 0 {
				env.Print(" ")
			}
			env.Print((*BinaryValue)(&p.Payload).String())
		}
		env.Println("")
	} else {
		env.Println("Authorization denied")
	}
}

type tagSet map[string]struct{}

type listParams struct {
	tags tagSet
}

func (p *listParams) Run(env CLIEnv) {
	if len(p.tags) == 0 {
		env.Println("(none)")
		return
	}

	list := []string{}
	for tag, _ := range p.tags {
		list = append(list, tag)
	}
	sort.Strings(list)
	env.Println(strings.Join(list, ", "))
}

type addParams struct {
	tags tagSet
	Name string `usage:"Tag name to add." cli:",arg"`
}

func (p *addParams) Run(env CLIEnv) {
	p.tags[p.Name] = struct{}{}
}

type dropParams struct {
	tags tagSet
	Name string `usage:"Tag name to drop." cli:",arg"`
}

func (p *dropParams) Run(env CLIEnv) {
	delete(p.tags, p.Name)
}

func main() {
	tags := tagSet{}

	LaunchOS("Testing the CLI mini-framework.", NewCLI("> ", true,
		&Command{"test", "Test argument parsing.", &params{Count: 1}},
		&Command{"tag", "Manage a set of tags.", NewCLI("tag> ", true,
			&Command{"list", "View all tags.", &listParams{tags: tags}},
			&Command{"add",  "Add a tag.",     &addParams{tags: tags}},
			&Command{"drop", "Remove a tag.",  &dropParams{tags: tags}},
		)},
	))
}
