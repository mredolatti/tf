# vi: ft=dockerfile
FROM golang:1.20.4-bullseye as builder

RUN apt update
RUN apt install -y bash build-essential git

WORKDIR /app
COPY . .

RUN make file-server
RUN make fsbasic.so

# --- runner stage

FROM debian:bullseye AS runner

#RUN apk add bash
#
## Support for libresolv (needed by plugins)
#RUN apk add gcompat
#RUN gcompat include libresolv.so.2

RUN apt update
RUN apt install -y netcat libc6

RUN mkdir -p /opt/mifs/plugins
RUN mkdir -p /var/mifs/files
RUN mkdir -p /var/mifs/authdb
RUN uname -a > /var/mifs/files/serverinfo.txt

COPY --from=builder /app/infra/waitport.sh /usr/local/bin
COPY --from=builder /app/infra/fs-entrypoint.sh /usr/local/bin
COPY --from=builder /app/file-server /usr/local/bin
COPY --from=builder /app/fsbasic.so /opt/mifs/plugins/

ENTRYPOINT ["fs-entrypoint.sh"]
