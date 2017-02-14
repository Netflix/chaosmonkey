To run unit tests:

```bash
go test ./...
```

## Tests that interact with MySQL

There are some tests that interact with MySQL. the test files are
`mysql/*_test.go`

These tests assume a MySQL deployment at the following connection string:

```
root:password@tcp(127.0.0.1:3306)/
```

### Testing with Docker

The simplest way to run these tests is to install Docker on your local machine.
These tests use the `mysql:5.6` container (version 5.6 is used to ensure
compatibility with [Amazon Aurora][1]).

Note that if you are on macOS, you must use [Docker for Mac][2], not Docker
Toolbox. Otherwise, the Docker containers will not be accessible at 127.0.0.1.


If you want to run these tests, ensure you have Docker installed locally, and
grab the mysql:5.6 container:

```bash
docker pull mysql:5.6
```

Then run the tests with the `docker` tag, like this:

```
go test -tags docker  ./...
```

The tests will automatically start the mysql container and then bring it down.

### Testing without bringing Docker container up and down

If you don't want the tests to bring the mysql Docker container up and down each
time (e.g., you want to run the tests more quickly, or you want to test by
running a mysql instance natively), use the "dockerup" flag along with the
"docker" flag.

```
go test -tags "docker dockerup"  ./...
```

(In retrospect, "docker" and "dockerup" are not great names for these tag, maybe "mysqltests"
and "nodocker" would be better).

[1]: https://aws.amazon.com/rds/aurora
[2]: https://docs.docker.com/engine/installation/mac/
