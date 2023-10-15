FROM alpine:latest as certs
RUN apk add --no-cache -U ca-certificates

FROM busybox:musl
ARG USER=auth-proxy
RUN adduser -D ${USER}
USER ${USER}

COPY auth-proxy /auth-proxy
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080
CMD ["/auth-proxy"]
