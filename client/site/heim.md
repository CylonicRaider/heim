# Euphoria is open-source

The software powering Euphoria is called Heim. Its source code is available
[on GitHub](https://github.com/CylonicRaider/heim).

The server is written in Go. The client is written in JavaScript using React
15.x. They communicate by exchanging JSON messages over WebSockets. The
client-server API is [documented](heim/api).

You may have found this page due to the string `euphoria.leet.nu/heim`
mentioned in the Go source code. The server can be installed by running
`go install euphoria.leet.nu/heim`, but will be largely non-functional without
the client.
