FROM busybox
COPY caas /usr/bin/
EXPOSE 80
ENTRYPOINT ["/usr/bin/caas"]
