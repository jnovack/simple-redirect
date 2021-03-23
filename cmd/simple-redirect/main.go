package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
	"time"

	version "github.com/jnovack/release"
	metrics "github.com/jnovack/simple-redirect/internal/metrics"

	"github.com/jnovack/flag"

	"github.com/mattn/go-isatty"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	promPort  = 3000
	httpPort  = 8080
	httpsPort = 8443
)

var (
	target   = flag.String("target", "http://localhost", "target for redirect")
	status   = flag.Int("status", 301, "http redirect status code")
	certfile = flag.String("certfile", "", "certificate (in pem format) for tls")
	keyfile  = flag.String("keyfile", "", "private key (in pem format) for tls")
)

func redirect(w http.ResponseWriter, req *http.Request) {

	if req.TLS == nil {
		metrics.HTTPRedirects++
	} else {
		metrics.HTTPSRedirects++
	}

	// TODO Optionally add Path
	if req.URL.Path != "/" {
		*target += req.URL.Path
	}

	// TODO Optionally add Query
	if len(req.URL.RawQuery) > 0 {
		*target += "?" + req.URL.RawQuery
	}

	// TODO Log
	// log.Printf("redirect to: %s", target)

	http.Redirect(w, req, *target,
		// 301 - Permanently Moved http.StatusMovedPermanently
		// 302 - Temporarily Moved http.StatusFound
		*status)

}

func main() {
	flag.Parse()

	// TODO Optionally set Target
	target := "https://www.google.com"

	// TODO Optionally set Status
	status := http.StatusFound

	// Set metrics
	metrics.Target = target
	metrics.Status = status

	// Serve HTTP redirects
	log.Info().Msg("Serving redirects on :" + strconv.FormatInt(int64(httpPort), 10))
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", redirect)
	go func() {
		log.Fatal().Err(http.ListenAndServe(":"+strconv.FormatInt(int64(httpPort), 10), httpMux)).Msg("Error with http.ListenAndServe()")
	}()

	// Serve HTTPS redirects
	// if you have a client certificate you want a key aswell
	if *certfile != "" && *keyfile != "" {
		_, err := tls.LoadX509KeyPair(*certfile, *keyfile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load certificate/keypair")
		}
		httpsMux := http.NewServeMux()
		httpsMux.HandleFunc("/", redirect)

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				// TLS v1.3
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				// TLS v1.2
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
		}

		https := &http.Server{
			Addr:         ":" + strconv.FormatInt(int64(httpsPort), 10),
			Handler:      httpsMux,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}

		log.Info().Msg("Serving redirects on :" + strconv.FormatInt(int64(httpsPort), 10))
		go func() {
			log.Fatal().Err(https.ListenAndServeTLS(*certfile, *keyfile)).Msg("Error with https.ListenAndServeTLS()")
		}()

	} else if (*certfile != "" && *keyfile == "") ||
		(*certfile == "" && *keyfile != "") {
		log.Warn().Msg("Warning: For TLS to work both certificate and private key are needed. Skipping TLS.")
	}

	// Serve Prometheus statistics
	prom := http.NewServeMux()
	prom.Handle("/metrics", promhttp.Handler())
	prom.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>` + version.Application + ` Exporter</title></head>
			<body>
			<h1>` + version.Application + ` Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Info().Msg("Serving metrics on :" + strconv.FormatInt(int64(promPort), 10))
	log.Fatal().Err(http.ListenAndServe(":"+strconv.FormatInt(int64(promPort), 10), prom)).Msg("Error with prom.ListenAndServe()")

}

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		// Format using ConsoleWriter if running straight
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().In(time.Local)
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		// Format using JSON if running as a service (or container)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	prometheus.MustRegister(metrics.NewCollector())
}
