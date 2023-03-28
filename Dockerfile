FROM --platform="${BUILDPLATFORM}" docker.io/library/busybox:1.36.0@sha256:b5d6fe0712636ceb7430189de28819e195e8966372edfc2d9409d79402a0dc16 AS picker

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

FROM --platform="${TARGETPLATFORM}" gcr.io/distroless/static:nonroot@sha256:149531e38c7e4554d4a6725d7d70593ef9f9881358809463800669ac89f3b0ec

LABEL \
  org.opencontainers.image.source="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.url="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

COPY --from=picker /pick/bitbucket-semantic-pull-requests /app

ENTRYPOINT ["/app/bitbucket-semantic-pull-requests"]
