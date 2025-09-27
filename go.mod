module simple_blockchain

go 1.24.0

toolchain go1.24.6

require github.com/go-chi/chi/v5 v5.2.3 // direct

require (
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/syndtr/goleveldb v1.0.0
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
