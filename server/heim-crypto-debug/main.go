package main

func main() {
	c := NewDefaultConsole()
	defer c.Close()

	name := c.ReadLine("Username: ")
	pass := c.ReadPass("Password: ")
	if pass == "hunter2" {
		c.Println("Hello " + name + "!")
	} else  {
		c.Println("Invalid username or password.")
	}
}
