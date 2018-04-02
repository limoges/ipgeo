
default: run

run:
	go run cmd/ipgeo.go \
		--locations data/locations.csv \
		--networks data/ipv4.csv \
		--port 8080 \
		--verbose
