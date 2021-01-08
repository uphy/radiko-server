FROM golang:1.16beta1 AS backend-builder

WORKDIR /work
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /radiko-server

FROM node:12-alpine AS frontend-builder

WORKDIR /work
COPY frontend/package.json .
COPY frontend/yarn.lock . 
RUN yarn install
COPY frontend/ ./
RUN yarn build

FROM alpine:3.12.3

RUN apk --no-cache add tzdata ffmpeg

WORKDIR /radiko-server/
COPY --from=backend-builder /radiko-server .
RUN chmod +x radiko-server
COPY --from=frontend-builder /work/dist static
RUN ls -al .

ENTRYPOINT [ "/radiko-server/radiko-server" ]