FROM golang:1.22-alpine3.20 as builder

RUN apk update && apk add git make bash build-base gcc bc

RUN mkdir -p /go/src/github.com/Netflix/chaosmonkey
WORKDIR /go/src/github.com/Netflix/chaosmonkey
ADD ./ /go/src/github.com/Netflix/chaosmonkey


RUN make all

FROM ubuntu:24.04

COPY --from=builder /go/src/github.com/Netflix/chaosmonkey/build/chaosmonkey /opt/chaosmonkey/bin/chaosmonkey

ADD ./docs/chaosmonkey.toml /etc/chaosmonkey.toml
ADD ./docs/chaosmonkey-terminate.sh /opt/chaosmonkey/bin/chaosmonkey-terminate.sh
ADD ./docs/chaosmonkey-schedule.sh /opt/chaosmonkey/bin/chaosmonkey-schedule.sh



RUN apt-get update
RUN apt-get install -y cron

ADD ./docs/chaosmonkey-cron /etc/cron.d/chaosmonkey-cron

ENTRYPOINT ["cron", "-f"]
