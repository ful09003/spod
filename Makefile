
nexpone:
		@docker run --rm -d --name nexpone -p 9100:9100 --pid="host" -v "/:/host:ro,rslave" quay.io/prometheus/node-exporter:latest --path.rootfs=/host
nexptwo:
		@docker run --rm -d --name nexptwo -p 9101:9100 --pid="host" -v "/:/host:ro,rslave" quay.io/prometheus/node-exporter:v0.16.0

# gr8 for testing due to historical naming changes in node_exporter
node_exporters: nexpone nexptwo

bin: test
		@go build -o output/spud main.go

vet:
		@go vet

test: vet
