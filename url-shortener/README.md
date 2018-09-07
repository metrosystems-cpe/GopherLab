# URL Shortener

A basic url shortener service that uses [redis-service](https://github.com/metrosystems-cpe/GopherLab/tree/master/redis-service) for key - value storage

Features:

- metrics : prometheus with opencensus.io
- traces  : planned but not done !
- web     : simple html with materialize css and js + axios (credits for the gopher picture goes to the owner)
- tests   : planed but not done !

## Service description

```bash
                           - path /                --> index.html
                         /
user --> url-shortener < - - path /s&url={url}     --> receive a short url
                         \
                           - path /r/{key}         --> 301 to original url
```

## Examples

```bash
# Method GET a URL shorted
curl http://localhost:8081/s?url=https://medium.com/metrosystemsro/gitops-with-weave-flux-40997e929254
{"url":"localhost:8081/r/2600343750"}

# use a short url
curl http://localhost:8081/r/2600343750
<a href="https://medium.com/metrosystemsro/gitops-with-weave-flux-40997e929254">Moved Permanently</a>.

```

## url-shortener Contributors

- [@ionutvilie](https://github.com/ionutvilie)
- [@bogdanb07](https://github.com/bogdanb07)