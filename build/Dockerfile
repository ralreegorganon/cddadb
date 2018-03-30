FROM alpine:3.7

RUN apk --no-cache add ca-certificates

ADD bin/linux/cddadb /bin/cddadb
ADD /migrations /migrations
ENV CDDADB_MIGRATIONS_PATH=file:///migrations

ENTRYPOINT ["cddadb"]