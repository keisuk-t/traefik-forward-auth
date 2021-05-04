FROM gcr.io/distroless/static:nonroot

COPY bin/traefik-forward-auth ./

ENTRYPOINT ["./traefik-forward-auth"]
