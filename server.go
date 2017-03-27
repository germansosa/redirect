/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// A webserver that only redirect all requests.
package main

import (
	"github.com/namsral/flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
}

func main() {
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "REDIRECT", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port number to listen.")
	dest := fs.String("dest", "http://127.0.0.1/", "Destination url.")
	code := fs.Int("code", 301, "Status code. The provided code should be in the 3xx range.")
	auri := fs.Bool("auri", false, "Append request URI to destination url.")
	fs.Parse(os.Args[1:])

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if *auri == true {
			http.Redirect(w, r, *dest + r.RequestURI, *code)
		} else {
			http.Redirect(w, r, *dest, *code)
		}

		duration := time.Now().Sub(start).Seconds() * 1e3

		proto := strconv.Itoa(r.ProtoMajor)
		proto = proto + "." + strconv.Itoa(r.ProtoMinor)

		requestCount.WithLabelValues(proto).Inc()
		requestDuration.WithLabelValues(proto).Observe(duration)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	http.Handle("/metrics", promhttp.Handler())
	// TODO: Use .Shutdown from Go 1.8
        fmt.Printf("Serving on port %d. Redirecting to %s with status code %d\n", *port, *dest, *code)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not start http server: %s\n", err)
		os.Exit(1)
	}
}
