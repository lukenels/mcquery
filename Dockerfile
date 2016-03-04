FROM golang
ADD . /go
RUN go install mcbot
ENTRYPOINT /go/bin/mcbot --port 80
EXPOSE 80
