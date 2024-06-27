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
ci: podman ci-upload-in-container

.PHONY: ci-upload-in-container
ci-upload-in-container:
	@$(MAKE) podman-in-container command="make ci-upload" podman_args="-e CI_JOB_TOKEN=$$CI_JOB_TOKEN -e CI_API_V4_URL=$$CI_API_V4_URL -e CI_PROJECT_ID=$$CI_PROJECT_ID"

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

podman_build_dockerfile_hash?=$(shell git hash-object docker/build.Dockerfile)
podman_build_repo?=aqrm.io/aquareum-tv/aquareum
podman_build_ref?=$(podman_build_repo):builder-$(podman_build_dockerfile_hash)
.PHONY: podman
podman: podman-build-builder
	$(MAKE) podman-in-container command="make all"

.PHONY: podman-build-builder
podman-build-builder:
	cd docker \
	&& podman build \
		--os=linux \
		-f build.Dockerfile \
		--layers \
		--cache-to $(podman_build_repo) \
		--cache-from $(podman_build_repo) \
		-t $(podman_build_ref) . \
	&& podman push $(podman_build_ref)

.PHONY: podman-build-builder-if-necessary
podman-build-builder-if-necessary:
	podman pull $(podman_build_ref) || $(MAKE) podman-build-builder

command=echo 'no command specified' && exit 1
podman_args?=
.PHONY: podman-build-builder
podman-in-container:
	@podman run $(podman_args) -v $$(pwd):$$(pwd) -w $$(pwd) --rm -it $(podman_build_ref) bash -euo pipefail -c "$(command)"

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
