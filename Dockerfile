FROM golang:1.15-alpine As builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go
COPY api/ api/
COPY pkg/  pkg/
COPY controllers/ controllers/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/manager main.go

RUN apk add unzip \
    &&  wget -q  https://releases.hashicorp.com/terraform/0.13.5/terraform_0.13.5_linux_amd64.zip  \
    && unzip terraform_0.13.5_linux_amd64.zip -d /workspace/bin \
    && rm -rf terraform_0.13.5_linux_amd64.zip

FROM alpine:3.13.2

RUN addgroup nonroot && \
    adduser -S -G nonroot nonroot

USER nonroot

WORKDIR /

COPY --chown=nonroot:nonroot --from=builder /workspace/bin/terraform /usr/local/bin
COPY --chown=nonroot:nonroot --from=builder /workspace/bin/manager /manager

ENTRYPOINT ["/manager"]
