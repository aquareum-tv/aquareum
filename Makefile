OUT_DIR?="bin"
$(shell mkdir -p $(OUT_DIR))

.PHONY: default
default: app node

VERSION?=$(shell ./util/version.sh)

.PHONY: version
version:
	@./util/version.sh

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
all: version install check app node-all-platforms android

.PHONY: ci
ci: all

.PHONY: ci
ci: ci-pull all

.PHONY: ci-pull
ci-pull:
	git fetch --tags

.PHONY: android
android: app
	cd ./packages/app/android \
	&& ./gradlew build \
	&& cd - \
	&& mv ./packages/app/android/app/build/outputs/apk/release/app-release.apk ./bin/aquareum-$(VERSION)-android-release.apk \
	&& mv ./packages/app/android/app/build/outputs/apk/debug/app-debug.apk ./bin/aquareum-$(VERSION)-android-debug.apk

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
			&& $(MAKE) ci-upload-file upload_file=$$file; \
		done \
	done; \
	$(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-android-release.apk \
	&& $(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-android-debug.apk

upload_file?=""
.PHONY: ci-upload-file
ci-upload-file:
	curl \
		--fail-with-body \
		--header "JOB-TOKEN: $$CI_JOB_TOKEN" \
		--upload-file bin/$(upload_file) \
		"$$CI_API_V4_URL/projects/$$CI_PROJECT_ID/packages/generic/aquareum/$(VERSION)/$(upload_file)";

.PHONY: check
check:
	yarn run check

.PHONY: fix
fix:
	yarn run fix
