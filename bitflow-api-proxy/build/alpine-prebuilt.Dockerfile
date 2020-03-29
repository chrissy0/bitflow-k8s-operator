# teambitflow/bitflow-api-proxy
# Copies pre-built binaries into the container. The binaries are built on the local machine beforehand:
# ./alpine-build.sh
# docker build -t teambitflow/bitflow-api-proxy -f alpine-prebuilt.Dockerfile _output/bin/alpine
FROM alpine:3.11.5
RUN apk --no-cache add libstdc++
COPY bitflow-api-proxy /
ENTRYPOINT ["/bitflow-api-proxy"]
