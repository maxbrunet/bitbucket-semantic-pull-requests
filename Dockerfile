FROM --platform="${BUILDPLATFORM}" docker.io/library/busybox:1.35.0@sha256:43eb0902ca9a91a2d37c342eba51cb9790e9314c3b5bdfca1cf49ebf2169a2d9 AS picker

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY dist /dist

RUN mkdir /pick && \
    if [ "${TARGETARCH}" == 'amd64' ]; then \
        # https://github.com/golang/go/wiki/MinimumRequirements#amd64
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT:-v1}/bitbucket-semantic-pull-requests" /pick; \
    elif [ "${TARGETARCH}" == 'arm' ]; then \
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT##v}/bitbucket-semantic-pull-requests" /pick; \
    else \
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}/bitbucket-semantic-pull-requests" /pick; \
    fi

FROM gcr.io/distroless/static:nonroot@sha256:ed05c7a5d67d6beebeba19c6b9082a5513d5f9c3e22a883b9dc73ec39ba41c04

LABEL \
  org.opencontainers.image.source="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.url="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

COPY --from=picker /pick/bitbucket-semantic-pull-requests /app

ENTRYPOINT ["/app/bitbucket-semantic-pull-requests"]
