FROM --platform="${BUILDPLATFORM}" docker.io/library/busybox:1.35.0@sha256:f4ed5f2163110c26d42741fdc92bd1710e118aed4edb19212548e8ca4e5fca22 AS picker

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

FROM gcr.io/distroless/static:nonroot@sha256:cacb3d85a593c44e2f3998175164f0c694e475b40e639294f224062842e90724

LABEL \
  org.opencontainers.image.source="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.url="https://github.com/maxbrunet/bitbucket-semantic-pull-requests" \
  org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

COPY --from=picker /pick/bitbucket-semantic-pull-requests /app

ENTRYPOINT ["/app/bitbucket-semantic-pull-requests"]
