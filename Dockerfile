FROM busybox:musl AS bin
COPY auth-proxy /auth-proxy
EXPOSE 8080
CMD ["/auth-proxy"]
