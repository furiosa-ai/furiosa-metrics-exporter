ARG BASE_IMAGE=asia-northeast3-docker.pkg.dev/next-gen-infra/furiosa-ai/libfuriosa-kubernetes:38c0ce0-test

FROM $BASE_IMAGE as build
ARG TARGETARCH

# Build metric-exporter binary
WORKDIR /
COPY . /
RUN make build

# Stage arch-specific runtime libs into a path tree mirroring /usr/lib/<triplet>/
# so the distroless final stage (no shell) can COPY them to the correct location.
RUN set -eux; \
    case "$TARGETARCH" in \
        amd64) libDir='x86_64-linux-gnu' ;; \
        arm64) libDir='aarch64-linux-gnu' ;; \
        *) echo >&2 "unsupported architecture: $TARGETARCH"; exit 1 ;; \
    esac; \
    mkdir -p /staging/usr/lib/$libDir; \
    cp /usr/lib/$libDir/libfuriosa_smi.so /staging/usr/lib/$libDir/; \
    cp /usr/lib/$libDir/libgcc_s.so.1   /staging/usr/lib/$libDir/

FROM gcr.io/distroless/base-debian12:latest

# Copy device plugin binary
WORKDIR /

# Below dynamic libraries are required due to `furiosa-smi` and Rust dependencies.
# Copied as a tree so each arch lands in its own /usr/lib/<triplet>/ dir.
COPY --from=build /staging/ /

COPY --from=build /main /main

CMD ["./main"]
