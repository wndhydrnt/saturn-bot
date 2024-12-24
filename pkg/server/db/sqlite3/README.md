This is a copy of the sqlite3 database driver [sqlite3.go](https://github.com/golang-migrate/migrate/blob/c378583d782e026f472dff657bfd088bf2510038/database/sqlite3/sqlite3.go) from the [golang-migrate](https://github.com/golang-migrate/migrate) project.

It makes golang-migrate compatible with [github.com/ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3)
and removes the need for CGO, which `github.com/mattn/go-sqlite3` would introduce.

golang-migrate is distributed under the MIT license.
A copy of the license can be viewed [here](./LICENSE).
