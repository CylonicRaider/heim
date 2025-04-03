# ◓

Heim is the backend and frontend of [euphoria](https://euphoria.leet.nu), a
real-time community platform. The backend is a Go server that speaks JSON over
WebSockets, persisting data to PostgreSQL. Our web client is built with
React 15.x and Reflux.

## Getting started

1. Install `git`, [`docker`](https://docs.docker.com/engine/install/), and
   [`docker-compose`](https://docs.docker.com/compose/install/).

2. Ensure dependencies are fetched: Run `git submodule update --init` in this
   repo directory.

### Running a server

1. Build the client static files: `docker-compose run --rm frontend`

2. Init your db: `docker-compose run --rm upgradedb sql-migrate up`

3. Start the server: `docker-compose up backend`

Heim is now running on port 8080. \o/

### Developing the client (connected to the main instance)

1. Launch the standalone static server and build watcher:  
   `docker-compose run --service-ports frontend gulp develop`

2. To connect to [&test](https://euphoria.leet.nu/room/test) on euphoria.leet.nu
   using your local client, open:  
   <http://localhost:8080/room/test/#connect=https://euphoria.leet.nu>

### Running tests

#### Backend

`docker-compose run --rm backend_tests`

#### Frontend

`docker-compose run --rm frontend npm test`

### Updating dependencies

`docker-compose run --rm update_deps`

## Self-hosting

The container set defined in `docker-compose.yml` may serve as a starting
point, but it is **not suitable for production use**. It includes and
references clear-text hard-coded passwords and cryptographic keys.

In order to run an Internet-facing Heim instance, you need at least:
- A PostgreSQL database. (This is *technically* optional, but you will
  regularly lose all data without a database.)
- A backend (`heimctl serve`). This needs to talk to the database using a
  strong random password.
- An embed server (`heimctl server-embed`). This does not require database
  access.
- Likely, the backend console (offered by instances of the backend). This
  requires strong SSH host and client keys.

## Discussion

Questions? Feedback? Ideas? Come join us in
[&heim](https://euphoria.leet.nu/room/heim) or email euphoria at leet.nu.

## Licensing

The server and the client are distributed under the terms of the
GNU Affero General Public License 3.0.

Art and documentation are distributed under the terms of the CC-BY 4.0 license.

See [LICENSE.md](LICENSE.md) for licensing details.
