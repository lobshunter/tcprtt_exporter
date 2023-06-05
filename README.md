# TCPRTT Exporter

TCPRTT Exporter is a toy that runs `ip tcp_metrics` command to [collect tcp rtt value](https://stackoverflow.com/a/59663237) from host, can be used to monitor network latency jitter. Note that since TCP source/destination IP is exported as metrics label, it could meet [high cardinality](https://blog.cloudflare.com/how-cloudflare-runs-prometheus-at-scale/) issue of prometheus.

## Usage

```bash
# start monitoring and will export metrics to http://0.0.0.0:8090/metrics
./tcprtt_exporter
```

```bash
# like command above, but will try to resolve ip address to pod/service name in kubernetes cluster
./tcprtt_exporter -kubeconfig $KUBECONFIG -resolve-kubernetes
```

## TODO

- example kubernetes deployment manifest
- option to collect rtt using eBPF, since data from `ip tcp_metrics` is only updated when TCP connection is closed.
