FROM golang:1.18.5-alpine AS backend
WORKDIR /go/src/github.com/broothie/slink.chat
COPY . .
RUN go build -o slink ./cmd/server/main.go
RUN go run ./cmd/bundle/main.go --production

FROM node:18.6.0 AS frontend
WORKDIR /usr/src/app
COPY . .
RUN yarn
RUN yarn css

FROM alpine:3.16.2
COPY --from=backend /go/src/github.com/broothie/slink.chat/slink slink
COPY --from=backend /go/src/github.com/broothie/slink.chat/static static
COPY --from=frontend /usr/src/app/static/style.css static/style.css
COPY templates templates
CMD ./slink
