FROM debian:11-slim

ARG arch amd64

RUN apt update && apt install -y \
    iproute2

COPY bin/linux/${arch}/tcprtt_exporter /usr/local/bin/tcprtt_exporter

ENTRYPOINT [ "/usr/local/bin/tcprtt_exporter" ]
