package main

import (
	"crypto/tls"
	goflag "flag"
	"fmt"
	"net/http"

	"k8s-validating-webhook/pkg/handler"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	port := flag.Int("port", 443, "Port to listen on")
	metricsPort := flag.Int("metrics-port", 8090, "Port to listen on for metrics and healthz (http)")
	//inCluster := flag.Bool("in-cluster", true, "Use in-cluster authentication and certificate request API")
	tlsKeyFile := flag.String("tls-key", "/etc/webhook/certs/tls.key", "(out-of-cluster) TLS key file path")
	tlsCertFile := flag.String("tls-cert", "/etc/webhook/certs/tls.crt", "(out-of-cluster) TLS certificate file path")

	klog.InitFlags(goflag.CommandLine)
	goflag.CommandLine.VisitAll(func(f *goflag.Flag) {
		flag.CommandLine.AddFlag(flag.PFlagFromGoFlag(f))
	})
	flag.Parse()
	_ = goflag.CommandLine.Parse([]string{})

	addr := fmt.Sprintf(":%d", *port)
	metricsAddr := fmt.Sprintf(":%d", *metricsPort)
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", handler.Validation)

	//for merics
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	// setup signal handler to be passed to certwatcher and http server

	signalHandlerCtx := signals.SetupSignalHandler()
	tlsConfig := &tls.Config{}

	//if !*inCluster {
	klog.Info("Cert Watching")
	watcher, err := certwatcher.New(*tlsCertFile, *tlsKeyFile)
	if err != nil {
		klog.Fatalf("Error initializing certwatcher: %q", err)
	}

	go func() {
		if err := watcher.Start(signalHandlerCtx); err != nil {
			klog.Fatalf("Error starting certwatcher: %q", err)
		}
	}()

	tlsConfig.GetCertificate = watcher.GetCertificate
	//}

	klog.Info("Creating server")
	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	metricsServer := &http.Server{
		Addr:    metricsAddr,
		Handler: metricsMux,
	}

	go func() {
		klog.Infof("Listening on %s", addr)
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			klog.Fatalf("Error listening: %q", err)
		}
	}()

	klog.Infof("Listening on %s for metrics and healthz", metricsAddr)
	if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
		klog.Fatalf("Error listening: %q", err)
	}

}
