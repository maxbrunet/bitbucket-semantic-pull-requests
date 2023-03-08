FROM --platform="${BUILDPLATFORM}" docker.io/library/busybox:1.36.0@sha256:c118f538365369207c12e5794c3cbfb7b042d950af590ae6c287ede74f29b7d4 AS picker

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY dist /dist

RUN mkdir /pick && \
    if [ "${TARGETARCH:-amd64}" = 'amd64' ]; then \
        # https://github.com/golang/go/wiki/MinimumRequirements#amd64
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS:-linux}_${TARGETARCH:-amd64}_${TARGETVARIANT:-v1}/bitbucket-semantic-pull-requests" /pick; \
    elif [ "${TARGETARCH}" = 'arm' ]; then \
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT##v}/bitbucket-semantic-pull-requests" /pick; \
    else \
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}/bitbucket-semantic-pull-requests" /pick; \
    fi

FROM --platform="${TARGETPLATFORM}" gcr.io/distroless/static:nonroot@sha256:775156d9ad597260d6e9cd0a8a400e56f2a860b845e54d8e48ce6da22f90240f

LABEL \
  org.opencontainers.image.source="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.url="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

COPY --from=picker /pick/bitbucket-semantic-pull-requests /app

ENTRYPOINT ["/app/bitbucket-semantic-pull-requests"]
