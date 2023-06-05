package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lobshunter/tcprtt_exporter/pkg/tcprttexporter"
	"github.com/lobshunter/tcprtt_exporter/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress     = flag.String("listen", ":8090", "address to bind http server to")
	resolveKubernetes = flag.Bool("resolve-kubernetes", false, "try resolving ip address to kubernetes service/pod name")
	kubeconfig        = flag.String("kubeconfig", "", "absolute path to the kubeconfig file, only required if out-of-cluster and resolve-kubernetes is true")
)

func main() {
	flag.Parse()

	collector := tcprttexporter.NewTCPRttCollector()
	if *resolveKubernetes {
		ipResolver, err := tcprttexporter.NewKubernetesIPResolver(*kubeconfig)
		if err != nil {
			log.Fatalf("[FATAL] create kubernetes ip resolver failed: %v", err)
		}
		collector.WithIPResolver(ipResolver)
	}

	r := prometheus.NewRegistry()
	err := r.Register(collector)
	if err != nil {
		log.Fatalf("[FATAL] register collector failed: %v", err)
	}

	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	go http.ListenAndServe(*listenAddress, nil)

	fmt.Println(version.Version())
	fmt.Println()
	log.Println("[INFO] server listening on", *listenAddress)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-shutdown
	log.Printf("[INFO] got signal %s, exiting", sig.String())
}
