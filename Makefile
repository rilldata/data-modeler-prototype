PACKAGE_NAME          := github.com/goreleaser/goreleaser-cross-example
GOLANG_CROSS_VERSION  ?= v1.19.2

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2

.PHONY: cli
cli:
	npm install
	npm run build
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/ cli/pkg/web/embed/dist
	# go build -o rill cli/main.go

# These commands not working in local, need to look into this.
# .PHONY: sysroot-pack
# sysroot-pack:
# 	@tar cf - $(SYSROOT_DIR) -P | pv -s $[$(du -sk $(SYSROOT_DIR) | awk '{print $1}') * 1024] | pbzip2 > $(SYSROOT_ARCHIVE)

# .PHONY: sysroot-unpack
# sysroot-unpack:
# 	@pv $(SYSROOT_ARCHIVE) | pbzip2 -cd | tar -xf -

.PHONY: release-dry-run
release-dry-run:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--rm-dist --skip-validate --skip-publish

# .PHONY: release
# release:
# 	@if [ ! -f ".release-env" ]; then \
# 		echo "\033[91m.release-env is required for release\033[0m";\
# 		exit 1;\
# 	fi
# 	docker run \
# 		--rm \
# 		--privileged \
# 		-e CGO_ENABLED=1 \
# 		--env-file .release-env \
# 		-v /var/run/docker.sock:/var/run/docker.sock \
# 		-v `pwd`:/go/src/$(PACKAGE_NAME) \
# 		-v `pwd`/sysroot:/sysroot \
# 		-w /go/src/$(PACKAGE_NAME) \
# 		goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
# 		release --rm-dist
