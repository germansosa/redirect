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
	"context"
	"fmt"
	"github.com/namsral/flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
	metricsPort := fs.Int("metrics-port", 9237, "Port number to serve metrics at /metrics.")
	redirectTo := fs.String("redirect-to", "", "Destination url.")
	statusCode := fs.Int("status-code", 301, "Status code. The provided code should be in the 3xx range.")
	appendURI := fs.Bool("append-uri", false, "Append request URI to destination url.")

	fs.Parse(os.Args[1:])

	if len(*redirectTo) == 0 {
		fmt.Println("-redirect-to is required.")
		os.Exit(1)
	}

	if *appendURI == true {
		// Remove trailing slash if exists
		*redirectTo = strings.TrimSuffix(*redirectTo, "/")
	}

	redirectMux := http.NewServeMux()
	redirectMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if *appendURI == true {
			http.Redirect(w, r, *redirectTo+r.RequestURI, *statusCode)
		} else {
			http.Redirect(w, r, *redirectTo, *statusCode)
		}

		duration := time.Now().Sub(start).Seconds() * 1e3

		proto := strconv.Itoa(r.ProtoMajor)
		proto = proto + "." + strconv.Itoa(r.ProtoMinor)

		requestCount.WithLabelValues(proto).Inc()
		requestDuration.WithLabelValues(proto).Observe(duration)
	})
	redirectMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	redirectSrv := &http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: redirectMux}
	metricsSrv := &http.Server{Addr: fmt.Sprintf(":%d", *metricsPort), Handler: metricsMux}

	go func() {
		// Main http server
		err := redirectSrv.ListenAndServe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not start http server: %s\n", err)
		}
	}()

	go func() {
		// Serving metrics
		err := metricsSrv.ListenAndServe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not start http server for metrics: %s\n", err)
		}
	}()

	fmt.Printf("Serving on port %d. Redirecting to %s with status code %d\n", *port, *redirectTo, *statusCode)

	// Wait for end signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM)
	// Block until SIGTERM is received.
	<-stopChan
	fmt.Println("Received SIGTERM, shutting down")

	// shut down gracefully, but wait no longer than 5 seconds before halting
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	redirectSrv.Shutdown(ctx)
	metricsSrv.Shutdown(ctx)
}
