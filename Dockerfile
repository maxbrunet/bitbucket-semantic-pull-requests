FROM --platform="${BUILDPLATFORM}" docker.io/library/busybox:1.35.0@sha256:20246233b52de844fa516f8c51234f1441e55e71ecdd1a1d91ebb252e1fd4603 AS picker

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY dist /dist

RUN mkdir /pick && \
    if [ "${TARGETARCH}" == 'arm' ]; then \
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}_${TARGETVARIANT##v}/bitbucket-semantic-pull-requests" /pick; \
    else \
        cp "/dist/bitbucket-semantic-pull-requests_${TARGETOS}_${TARGETARCH}/bitbucket-semantic-pull-requests" /pick; \
    fi

FROM gcr.io/distroless/static:nonroot@sha256:80c956fb0836a17a565c43a4026c9c80b2013c83bea09f74fa4da195a59b7a99

LABEL \
  org.opencontainers.image.source="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.url="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

COPY --from=picker /pick/bitbucket-semantic-pull-requests /app

ENTRYPOINT ["/app/bitbucket-semantic-pull-requests"]
