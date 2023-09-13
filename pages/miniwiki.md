*miniwiki* is a minimalistic wiki based on Markdown syntax, intended for personal use.
It is heavily inspired by [WikiMind](https://mindwi.se/?WikiMind), however written from scratch in Go.

*miniwiki* renders wiki pages in CommonMark and Github Flavored Markdown syntax as HTML (see [[demo]]) and allows page editing using a password.
It is licensed under the MIT license.

### How to use

The code is hosted at [github.com/lbrocke/miniwiki](https://github.com/lbrocke/miniwiki).

Program arguments:
```
  -dir string
        Directory of pages files (default "./pages/")
  -name string
        Wiki name (default "wiki")
  -addr string
        Listen host/port of web server
        (default 127.0.0.1:8080)
```

To enable page editing, set the `PASS` environment variable to some good password.

### Deploy

Either using Docker:

```
$ docker build -t miniwiki .
$ docker run \
    -e "PASS=changeme" \
    -e "NAME=mywiki" \
    -p 127.0.0.1:8080:80 \
    -v ./pages:/pages \
    miniwiki
```

or from source:

```
$ go install
$ PASS=changeme miniwiki \
    -name "mywiki" \
    -addr 127.0.0.1:8080 \
    -dir ./pages
```

Then put miniwiki behind some reverse proxy, e.g. [https://caddyserver.com](Caddy):

```
wiki.example.com {
      reverse_proxy 127.0.0.1:8080
}
```
