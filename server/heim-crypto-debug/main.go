package main

type params struct {
	Name     string `usage:"user name"`
	Password string `usage:"password (as plain text)"`
	Count    int    `usage:"repetition count"`
}

func (p *params) Run(con Console, argv []string) {
	if p.Name == "" {
		p.Name = *con.ReadLine("Username: ")
	}
	if p.Password == "" {
		p.Password = *con.ReadPass("Password: ")
	}

	if p.Password == "hunter2" {
		for i := 0; i < p.Count; i++ {
			con.Println("Hello " + p.Name + "!")
		}
	} else {
		con.Println("Authorization denied")
	}
}

func main() {
	con := NewDefaultConsole()
	defer con.Close()

	cli := NewCLI("> ")
	cli.Commands.AddNew("test", &params{Count: 1})
	cli.Run(con)
}
