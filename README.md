# redirect

A simple http server that redirect all requests.

## Usage

```bash
Usage of ./server:
  -append-uri
    	Append request URI to destination url.
  -metrics-port int
    	Port number to serve metrics at /metrics. (default 9237)
  -port int
    	Port number to listen. (default 8080)
  -redirect-to string
    	Destination url.
  -status-code int
    	Status code. The provided code should be in the 3xx range. (default 301)

$ ./server -redirect-to http://example.com -append-uri true
```
Parameters can also be used as environment variables with 'REDIRECT_' as prefix.
Example:
```bash
REDIRECT_REDIRECT_TO=http://example.com REDIRECT_APPEND_URI=true ./server
```

## Metrics

Metrics in prometheus format are served in /metrics on port 9237 by default (can be changed with -metrics-port)

## Docker

A minimalist image is pushed to docker hub: https://hub.docker.com/r/gsosa/redirect/
