.PHONY: cli
cli:
	npm install
	npm run build
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/ cli/pkg/web/embed/dist
	mkdir -p cli/pkg/examples
	cp -r examples/rill_example cli/pkg/examples
	go build -o rill cli/main.go 

.PHONY: proto.generate
proto.generate:
	cd proto && buf generate
