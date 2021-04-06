FROM busybox:musl AS bin
COPY auth-proxy /auth-proxy
ADD ./static /static
EXPOSE 8080
CMD ["/auth-proxy"]
