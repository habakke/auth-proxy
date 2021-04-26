FROM alpine:latest as certs
RUN apk add --no-cache -U ca-certificates

FROM busybox:musl
COPY auth-proxy /auth-proxy
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ADD static /static
ADD templates /templates
EXPOSE 8080
CMD ["/auth-proxy"]
