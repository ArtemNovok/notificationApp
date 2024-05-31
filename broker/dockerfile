FROM  golang:1.20-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY . /app

ENV APP_PASSOWRD=blank

RUN CGO_ENABLED=0 go build -o senderApp ./cmd/api

RUN chmod +x /app/senderApp

FROM alpine

RUN mkdir /app

WORKDIR /app

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip

ENV ZONEINFO /zoneinfo.zip

COPY --from=build /app/senderApp  /app

CMD [ "/app/senderApp" ]