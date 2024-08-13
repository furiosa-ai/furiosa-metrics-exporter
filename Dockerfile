FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest as build

# Build metric-exporter binary
WORKDIR /
COPY . /
RUN make build

FROM registry.corp.furiosa.ai/furiosa/libfuriosa-kubernetes:latest

# Copy metric-exporter binary
WORKDIR /
COPY --from=build /main /main
CMD ["./main"]
