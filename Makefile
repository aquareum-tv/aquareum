OUT_DIR?="bin"
$(shell mkdir -p $(OUT_DIR))
VERSION=$(shell git describe --long --tags --dirty | sed 's/-[0-9]*-g/-/')

.PHONY: default
default: app node

.PHONY: install
install:
	yarn install

.PHONY: app
app: install
	yarn run build

.PHONY: node
node:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o $(OUT_DIR)/aquareum ./cmd/aquareum

.PHONY: all
all: install check app node-all-platforms

.PHONY: node-all-platforms
node-all-platforms:
	for GOOS in linux; do \
		for GOARCH in amd64 arm64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH $(MAKE) node OUT_DIR=bin/$$GOOS-$$GOARCH \
			&& cd bin/$$GOOS-$$GOARCH \
			&& tar -czvf ../aquareum-$(VERSION)-$$GOOS-$$GOARCH.tar.gz ./aquareum \
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

.PHONY: ci-upload
ci-upload:
	for GOOS in linux; do \
		for GOARCH in amd64 arm64; do \
			export file=aquareum-$(VERSION)-$$GOOS-$$GOARCH.tar.gz \
			&& curl --fail-with-body --header "JOB-TOKEN: $$CI_JOB_TOKEN" --upload-file bin/$$file "$$CI_API_V4_URL/projects/$$CI_PROJECT_ID/packages/generic/aquareum/$(VERSION)/$$file"; \
		done \
	done

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: check
check:
	yarn run check

.PHONY: fix
fix:
	yarn run fix
