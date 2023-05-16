package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	_ = iota
	reSrcIndex
	reRttIndex
	reRttVarIndex
	reDstIndex
)

var (
	// match: "127.0.0.1 age 1748496.418sec cwnd 10 rtt 210409us rttvar 210409us source 10.233.75.119"
	rePattern = regexp.MustCompile(`(\S+) age .* rtt (\d+)us rttvar (\d+)us source (\S+)`)
)

var (
	hostName = getHostName()

	tcpRTT = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcprtt",
		Help: "tcp rtt in microsecond",
	}, []string{"src", "dst", "host"})

	tcpRTTVar = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcprttvar",
		Help: "tcp rtt variation in microsecond",
	}, []string{"src", "dst", "host"})
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	updateMetrics()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8090", nil)

	ticker := time.NewTicker(5 * time.Second)
	log.Printf("[INFO] server started")
	for {
		select {
		case <-ticker.C:
			updateMetrics()
		case sig := <-shutdown:
			log.Printf("got signal %s, exiting", sig.String())
			os.Exit(0)
		}
	}
}

func updateMetrics() {
	r, err := getStat()
	if err != nil {
		log.Printf("[ERROR] get stat failed: %v", err)
		return
	}

	buf := bytes.NewBuffer(nil)
	_, _ = buf.ReadFrom(r)
	lines := buf.String()
	for _, line := range rePattern.FindAllStringSubmatch(lines, -1) {
		src := line[reSrcIndex]
		dst := line[reDstIndex]
		rttString := line[reRttIndex]
		rtt, err := strconv.Atoi(rttString)
		if err != nil {
			log.Printf("[ERROR] convert rtt to int failed: %v", err)
			continue
		}

		rttVarString := line[reRttVarIndex]
		rttVar, err := strconv.Atoi(rttVarString)
		if err != nil {
			log.Printf("[ERROR] convert rttVar to int failed: %v", err)
			continue
		}

		tcpRTT.WithLabelValues(src, dst, hostName).Set(float64(rtt))
		tcpRTTVar.WithLabelValues(src, dst, hostName).Set(float64(rttVar))
	}
}

func getStat() (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	ssCmd := exec.Command("ip", "tcp_metrics")
	ssCmd.Stdout = buf
	ssCmd.Stderr = os.Stderr
	return buf, ssCmd.Run()
}

func getHostName() string {
	hostname, err := os.Hostname()
	must(err)
	return hostname
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
