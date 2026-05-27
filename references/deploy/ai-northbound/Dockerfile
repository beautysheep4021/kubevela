FROM golang:1.23.8-alpine AS builder

WORKDIR /workspace
COPY go.mod go.sum ./
ARG GOPROXY=https://proxy.golang.org
ENV GOPROXY=${GOPROXY}
RUN go mod download

COPY pkg/ pkg/
COPY references/cmd/ai-northbound/ references/cmd/ai-northbound/

ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/ai-northbound ./references/cmd/ai-northbound

FROM scratch
COPY --from=builder /out/ai-northbound /usr/local/bin/ai-northbound

USER 65532:65532
EXPOSE 8088
ENTRYPOINT ["/usr/local/bin/ai-northbound"]
CMD ["-addr", "0.0.0.0:8088"]
