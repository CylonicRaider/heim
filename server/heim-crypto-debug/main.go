package main

type params struct {
	Name     string `usage:"user name"`
	Password string `usage:"password (as plain text)"`
	Payload  []byte `usage:"binary payload" cli:",arg,required"`
	Count    int    `usage:"repetition count" cli:",arg"`
}

func (p *params) Run(con Console) {
	if p.Name == "" {
		p.Name = *con.ReadLine("Username: ")
	}
	if p.Password == "" {
		p.Password = *con.ReadPass("Password: ")
	}

	if p.Password == "hunter2" {
		con.Println("Hello " + p.Name + "!")
		for i := 0; i < p.Count; i++ {
			if i != 0 {
				con.Print(" ")
			}
			con.Print((*BinaryValue)(&p.Payload).String())
		}
		con.Println("")
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
