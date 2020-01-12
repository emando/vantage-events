FROM alpine:3.11
ADD ./dist/aggregator-linux-amd64 /aggregator
EXPOSE 443
ENTRYPOINT ["/aggregator", "start"]
