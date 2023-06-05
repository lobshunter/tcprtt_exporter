package tcprttexporter

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
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

var _ prometheus.Collector = &TCPRttCollector{}

type TCPRttCollector struct {
	hostname string

	tcprtt    *prometheus.Desc
	tcprttVar *prometheus.Desc

	ipResolver IPResolver
}

func NewTCPRttCollector() *TCPRttCollector {
	return &TCPRttCollector{
		hostname: getHostName(),

		tcprtt: prometheus.NewDesc(
			"tcprtt",
			"tcp rtt in microsecond",
			[]string{"src", "dst", "host"},
			nil,
		),
		tcprttVar: prometheus.NewDesc(
			"tcprttvar",
			"tcp rtt variation in microsecond",
			[]string{"src", "dst", "host"},
			nil,
		),
	}
}

func (t *TCPRttCollector) WithIPResolver(resolver IPResolver) *TCPRttCollector {
	t.ipResolver = resolver
	return t
}

func (t *TCPRttCollector) Describe(chan<- *prometheus.Desc) {}

func (t *TCPRttCollector) Collect(ch chan<- prometheus.Metric) {
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

		if t.ipResolver != nil {
			src = t.ipResolver.Resolve(src)
			dst = t.ipResolver.Resolve(dst)
		}

		ch <- prometheus.MustNewConstMetric(t.tcprtt, prometheus.GaugeValue, float64(rtt), src, dst, t.hostname)
		ch <- prometheus.MustNewConstMetric(t.tcprttVar, prometheus.GaugeValue, float64(rttVar), src, dst, t.hostname)
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
	if err != nil {
		hostname = "unknown"
	}
	return hostname
}
