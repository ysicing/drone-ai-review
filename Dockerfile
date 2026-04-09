FROM ysicing/god AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /go/src/

COPY go.mod go.mod

COPY go.sum go.sum

RUN go mod download

COPY . .

ARG GOOS=linux

ARG GOARCH=amd64

ARG CGO_ENABLED=0

RUN go build -o release/linux/amd64/plugin

FROM ysicing/debian

COPY --from=builder /go/src/release/linux/amd64/plugin /bin/drone-plugin

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh /bin/drone-plugin

ENTRYPOINT ["/entrypoint.sh"]

CMD [ "/bin/drone-plugin" ]
