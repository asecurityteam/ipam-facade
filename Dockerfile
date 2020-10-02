FROM asecurityteam/sdcli:v1 AS BUILDER
RUN mkdir -p /go/src/github.com/asecurityteam/ipam-facade
WORKDIR $GOPATH/src/github.com/asecurityteam/ipam-facade
COPY --chown=sdcli:sdcli . .
RUN sdcli go dep
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN packr2
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /opt/app main.go
# ensure we do not leak uid/guid that is only valid in this container to other steps
USER 0:0
RUN chown 0:0 /opt/app
RUN chmod 755 /opt/app

##################################

FROM alpine:latest as CERTS
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -r -0 /zoneinfo.zip .

###################################

FROM scratch
COPY --from=BUILDER /opt/app .
# the timezone data:
COPY --from=CERTS /zoneinfo.zip /
# the tls certificates:
COPY --from=CERTS /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV ZONEINFO /zoneinfo.zip
ENTRYPOINT ["/app"]
