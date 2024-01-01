# ◓

Heim is the backend and frontend of [euphoria](https://euphoria.leet.nu), a
real-time community platform. The backend is a Go server that speaks JSON over
WebSockets, persisting data to PostgreSQL. Our web client is built in
React/Reflux.

**Currently, heim is released in a pre-alpha state**. Please be advised that
new development is currently being prioritized over stability. We're releasing
in this form because we want to open up our codebase and development progress.
We will make breaking changes to the protocol, and will be slow to merge
complex pull requests while we get our core building blocks in place.

## Getting started

1. Install `git`, [`docker`](https://docs.docker.com/installation/), and
   [`docker-compose`](https://docs.docker.com/compose/install/).

2. Ensure dependencies are fetched: run `git submodule update --init` in this repo directory.

### Running a server

1. Build the client static files: `docker-compose run frontend`.

2. Init your db: `docker-compose run upgradedb sql-migrate up`.

3. Start the server: `docker-compose up backend`.

Heim is now running on port 8080. \o/

### Developing the client (connected to the main instance)

1. Launch the standalone static server and build watcher:  
   `docker-compose run --service-ports frontend gulp develop`

2. To connect to [&test](https://euphoria.leet.nu/room/test) on euphoria.leet.nu
   using your local client, open:
   <http://localhost:8080/room/test/#connect=https://euphoria.leet.nu>

### Running tests

#### Backend

`docker-compose run backend go test -v euphoria.io/heim/...`

#### Frontend

`docker-compose run frontend npm test`

## Discussion

Questions? Feedback? Ideas? Come join us in
[&heim](https://euphoria.leet.nu/room/heim) or email euphoria at leet.nu.

## Licensing

The server and the client are distributed under the terms of the
GNU Affero General Public License 3.0.

Art and documentation are distributed under the terms of the CC-BY 4.0 license.

See [LICENSE.md](LICENSE.md) for licensing details.
