FROM golang:1.18.5-alpine AS backend
WORKDIR /go/src/github.com/broothie/slink.chat
COPY . .
RUN go build -o job ./cmd/job/main.go

FROM alpine:3.16.2
COPY --from=backend /go/src/github.com/broothie/slink.chat/job job
CMD ./job
