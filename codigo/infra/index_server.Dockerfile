# vi: ft=dockerfile
FROM golang:1.20.4-alpine3.18 as builder

RUN apk add \
  bash \
  build-base \
  git

WORKDIR /app
COPY . .

RUN make index-server

# --- runner stage

FROM alpine:3.18 AS runner

RUN apk add bash

COPY --from=builder /app/infra/waitport.sh /usr/local/bin
COPY --from=builder /app/infra/is-entrypoint.sh /usr/local/bin
COPY --from=builder /app/index-server /usr/local/bin

ENTRYPOINT ["is-entrypoint.sh"]
