FROM --platform=$BUILDPLATFORM docker.io/library/golang:1-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN --mount=type=bind,target=/mnt /mnt/build.sh -o /server

FROM scratch
COPY --link --from=build /server /
ENTRYPOINT ["/server"]
