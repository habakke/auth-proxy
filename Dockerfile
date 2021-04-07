FROM alpine as certs
RUN apk update && apk add ca-certificates

FROM busybox:musl
COPY auth-proxy /auth-proxy
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
ADD static /static
EXPOSE 8080
CMD ["/auth-proxy"]
