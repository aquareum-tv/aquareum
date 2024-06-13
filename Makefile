OUT_DIR?="bin"
$(shell mkdir -p $(OUT_DIR))

.PHONY: default
default: app node

.PHONY: app
app:
	pnpm install
	pnpm run -r build

.PHONY: node
node:
	go build -o $(OUT_DIR)/aquareum ./cmd/aquareum

.PHONY: all
all: app node-all-platforms

.PHONY: node-all-platforms
node-all-platforms:
	for GOOS in linux; do \
		for GOARCH in amd64 arm64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH $(MAKE) node OUT_DIR=bin/$$GOOS-$$GOARCH \
			&& cd bin/$$GOOS-$$GOARCH \
			&& tar -czvf ../aquareum-$$GOOS-$$GOARCH.tar.gz ./aquareum \
			&& cd -; \
		done \
	done

.PHONY: docker-build
docker-build: docker-build-builder docker-build-in-container

.PHONY: docker-build-builder
docker-build-builder:
	cd docker \
	&& docker build --os=linux --arch=amd64 -f build.Dockerfile -t aqrm.io/aquareum-tv/aquareum:builder .

.PHONY: docker-build-builder
docker-build-in-container:
	docker run -v $$(pwd):$$(pwd) -w $$(pwd) --rm -it aqrm.io/aquareum-tv/aquareum:builder make
