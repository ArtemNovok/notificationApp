FROM  golang:1.20-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o emailsenderApp ./cmd/api

RUN chmod +x /app/emailsenderApp

FROM alpine

RUN mkdir /app

WORKDIR /app

COPY --from=build /app/emailsenderApp  /app

CMD [ "/app/emailsenderApp" ]