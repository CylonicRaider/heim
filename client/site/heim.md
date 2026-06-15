# Euphoria is open-source

The software powering Euphoria is called Heim. Its source code is available
[on GitHub](https://github.com/CylonicRaider/heim).

The server is written in Go. The client is written in JavaScript using React
15.x. They communicate by exchanging JSON messages over WebSockets. The
client-server API is [documented](/heim/api).

You may have found this page due to the string `euphoria.leet.nu/heim`
mentioned in the Go source code. That is the name of the Go module containing
the Heim server. It includes a few command-line programs that can be
installed via `go install euphoria.leet.nu/heim/NAME@latest`:

 * `heimctl`: The Heim server, including the Web/API server. Largely
   non-functional without a compiled client.

 * `heimlich`: A tool for packaging a Go binary and associated data files into
   a self-extracting ZIP file.

 * `heimflake`: A tool for inspecting "snowflake" IDs used for Heim messages
   and accounts.
