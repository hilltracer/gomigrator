FROM golang:1.23 AS build

ENV BIN_FILE /opt/gomigrator/gomigrator
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/gomigrator/*

FROM alpine:3.9

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="gomigrator"
LABEL MAINTAINERS="hilltracer@yandex.ru"

ENV BIN_FILE "/opt/gomigrator/gomigrator"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/gomigrator/config.yaml
COPY ./configs/config.yaml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
