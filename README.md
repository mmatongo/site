# dnlm.pw

Code for [dnlm.pw](https://dnlm.pw), a personal website.

## Technology

- HTML, CSS
- [GoLang](https://golang.org/)

## Development

```
$ go run ./cmd/dnlm
```

or

```
$ go build -o ./bin/web ./cmd/dnlm
$ ./bin/dnlm
```

The application will be available at [http://localhost:4000](http://localhost:4000).

The app also accepts an optional `-config` flag to specify the location of the configuration file. The default location is `./config/config.yml` so you can run the app with a custom configuration file like so:

```
$ go run ./cmd/dnlm -config /path/to/config.yml
```

By default, the app will look for a `config.yml` file in the `./config` directory. It will fail if it does not find one.

## License

The following directories and their contents are Copyright Daniel M. Matongo. You may not reuse anything therein without my permission:

```sh
blog/
notes/
ui/static/images/
```

All other directories and files are MIT Licensed (where applicable).

## TODO

- [] Footer
- [] Notes
- [] Recent Posts
- [x] Publish Date
- [] Tags
- [] Asset [compression](https://github.com/tdewolff/minify)
