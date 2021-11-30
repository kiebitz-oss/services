FROM golang:1.16 as builder 

WORKDIR /go/src/app
COPY * .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install ./...

FROM scratch

ENV KIEBITZ_SETTINGS=/settings

CMD [ "/kiebitz","run","all" ]
#CMD ["sleep","86400"]

COPY --from=builder /go/bin/kiebitz kiebitz

# Ports
EXPOSE 11111
EXPOSE 22222
