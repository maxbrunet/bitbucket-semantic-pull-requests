FROM --platform="${BUILDPLATFORM}" docker.io/library/busybox:1.35.0@sha256:c1bf51d77cd905198d1d00eee3438559847a14c22240859eaf4a97f87b1e5f2c AS picker

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

FROM gcr.io/distroless/static:nonroot@sha256:1f580b0a1922c3e54ae15b0758b5747b260bd99d39d40c2edb3e7f6e2452298b

LABEL \
  org.opencontainers.image.source="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.url="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

COPY --from=picker /pick/bitbucket-semantic-pull-requests /app

ENTRYPOINT ["/app/bitbucket-semantic-pull-requests"]
