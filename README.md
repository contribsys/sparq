# Sparq ⚡️

Sparq is a Go-based server which aims to implement the ActivityPub protocol, meaning it can be an "instance" within the Fediverse, along with compatibility with the Mastodon client API.

How's this different from Mastodon? Sparq is meant to be far more efficient with machine resources. Go programs typically take 1/10th of the RAM. Unlike Ruby, Go does not have a global thread lock; Sparq can automatically scale to use all CPUs on a machine.

## Contributing

**Sparq is under active development. It is NOT ready for use at this time.**

If you want to help with development, please join us in the issue tracker.

## Internals

Sparq is designed to be easy to deploy.
It is a single binary.
It does not require a separate Postgresql service.
It does not require a separate Sidekiq service.
Everything is started and managed internally by Sparq.
Where possible, Sparq uses the Go standard library.

* Database - pure-Go SQLite3. All data stored in a single file which can be backed up with a single `cp sparq.db ...`.
* Background Jobs - Sparq runs [Faktory](https://github.com/contribsys/faktory) and a pool of worker goroutines internally.
* Redis - started by Sparq as a child process. Nothing to manage.
* Media processing - uses ImageMagick command line tools like `convert` and `identify`.

## License

AGPL 3.0. Certain parts of the codebase are vendored dependencies and contain the relevant license.

## Author

[@getajobmike@ruby.social](https://ruby.social/@getajobmike)