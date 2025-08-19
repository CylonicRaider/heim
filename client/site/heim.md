# Euphoria is open-source

The software powering Euphoria is called Heim. Its source code is available
[on GitHub](https://github.com/CylonicRaider/heim).

The server is written in Go. The client is written in JavaScript using React
15.x. They communicate by exchanging JSON messages over WebSockets. The
client-server API is [documented](heim/api).

You may have found this page due to the string `euphoria.leet.nu/heim`
mentioned in the Go source code. Unfortunately, you cannot install the backend
via `go install euphoria.leet.nu/heim` (due to a missing feature in the Go
source code discovery scheme). Even if you could, you would have to retrieve
and build the client separately for the backend to be usable.
