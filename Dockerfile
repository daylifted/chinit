FROM traefik/whoami AS whoami

FROM scratch
COPY sh /bin/sh
COPY --from=whoami /whoami /whoami

HEALTHCHECK --interval=30s --timeout=5s \
  CMD ["/bin/sh", "http://localhost/", "--grep", "RemoteAddr"]

ENTRYPOINT ["/whoami"]