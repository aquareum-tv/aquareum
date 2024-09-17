OUT_DIR?="bin"
$(shell mkdir -p $(OUT_DIR))

.PHONY: default
default: app node

VERSION?=$(shell go run ./pkg/config/git/git.go -v)
UUID?=$(shell go run ./pkg/config/uuid/uuid.go)
BRANCH?=$(shell go run ./pkg/config/git/git.go --branch)

BUILDOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
BUILDARCH ?= $(shell uname -m | tr '[:upper:]' '[:lower:]')
ifeq ($(BUILDARCH),aarch64)
		BUILDARCH=arm64
endif
ifeq ($(BUILDARCH),x86_64)
		BUILDARCH=amd64
endif

.PHONY: version
version:
	@go run ./pkg/config/git/git.go -v \
	&& go run ./pkg/config/git/git.go --env -o .ci/build.env

.PHONY: install
install:
	yarn install --inline-builds

.PHONY: app-and-node
app-and-node:
	$(MAKE) app
	$(MAKE) node

.PHONY: app
app: schema install
	yarn run build

.PHONY: node
node: schema
	$(MAKE) meson-setup
	meson compile -C build aquareum
	mv ./build/aquareum ./bin/aquareum

.PHONY: schema
schema:
	mkdir -p js/app/generated \
	&& go run pkg/crypto/signers/eip712/export-schema/export-schema.go > js/app/generated/eip712-schema.json

.PHONY: test
test:
	meson test -C build go-tests

# test to make sure we haven't added any more dynamic dependencies
.PHONY: link-test
link-test:
	count=$(shell ldd ./build-linux-amd64/aquareum | wc -l) \
	&& echo $$count \
	&& if [ "$$count" != "6" ]; then echo "ldd reports new libaries linked! want 6 got $$count" \
		&& ldd ./bin/aquareum \
		&& exit 1; \
	fi

.PHONY: all
all: version install check app test node-all-platforms android

.PHONY: ci
ci: version install check app node-all-platforms ci-upload-node

.PHONY: ci-macos
ci-macos: version install check app node-all-platforms-macos ci-upload-node-macos ios ci-upload-ios

.PHONY: ci-macos
ci-android: version install check android ci-upload-android

.PHONY: ci-test
ci-test: app
	meson setup build $(OPTS)
	meson test -C build go-tests

.PHONY: android
android: app .build/bundletool.jar
	export NODE_ENV=production \
	&& cd ./js/app/android \
	&& ./gradlew :app:bundleRelease \
	&& ./gradlew :app:bundleDebug \
	&& cd - \
	&& mv ./js/app/android/app/build/outputs/bundle/release/app-release.aab ./bin/aquareum-$(VERSION)-android-release.aab \
	&& mv ./js/app/android/app/build/outputs/bundle/debug/app-debug.aab ./bin/aquareum-$(VERSION)-android-debug.aab \
	&& cd bin \
	&& java -jar ../.build/bundletool.jar build-apks --ks ../my-release-key.keystore --ks-key-alias alias_name --ks-pass pass:aquareum --bundle=aquareum-$(VERSION)-android-release.aab --output=aquareum-$(VERSION)-android-release.apks --mode=universal \
	&& java -jar ../.build/bundletool.jar build-apks --ks ../my-release-key.keystore --ks-key-alias alias_name --ks-pass pass:aquareum --bundle=aquareum-$(VERSION)-android-debug.aab --output=aquareum-$(VERSION)-android-debug.apks --mode=universal \
	&& unzip aquareum-$(VERSION)-android-release.apks && mv universal.apk aquareum-$(VERSION)-android-release.apk && rm toc.pb \
	&& unzip aquareum-$(VERSION)-android-debug.apks && mv universal.apk aquareum-$(VERSION)-android-debug.apk && rm toc.pb

.PHONY: ios
ios: app
	xcodebuild \
		-workspace ./js/app/ios/Aquareum.xcworkspace \
		-sdk iphoneos \
		-configuration Release \
		-scheme Aquareum \
		-archivePath ./bin/aquareum-$(VERSION)-ios-release.xcarchive \
		CODE_SIGN_IDENTITY=- \
		AD_HOC_CODE_SIGNING_ALLOWED=YES \
		CODE_SIGN_STYLE=Automatic \
		DEVELOPMENT_TEAM=ZZZZZZZZZZ \
		clean archive | xcpretty \
	&& cd bin \
	&& tar -czvf aquareum-$(VERSION)-ios-release.xcarchive.tar.gz aquareum-$(VERSION)-ios-release.xcarchive

# xcodebuild -exportArchive -archivePath ./bin/aquareum-$(VERSION)-ios-release.xcarchive -exportOptionsPlist ./js/app/exportOptions.plist -exportPath ./bin/aquareum-$(VERSION)-ios-release.ipa

.build/bundletool.jar:
	mkdir -p .build \
	&& curl -L -o ./.build/bundletool.jar https://github.com/google/bundletool/releases/download/1.17.0/bundletool-all-1.17.0.jar

OPTS = -D "gst-plugins-base:audioresample=enabled" \
		-D "gst-plugins-base:playback=enabled" \
		-D "gst-plugins-base:opus=enabled" \
		-D "gst-plugins-base:gio-typefinder=enabled" \
		-D "gst-plugins-base:typefind=enabled" \
		-D "gst-plugins-good:matroska=enabled" \
		-D "gst-plugins-bad:fdkaac=enabled" \
		-D "gstreamer-full:gst-full=enabled" \
		-D "gstreamer-full:gst-full-plugins=libgstaudioresample.a;libgstmatroska.a;libgstfdkaac.a;libgstopus.a;libgstplayback.a;libgsttypefindfunctions.a" \
		-D "gstreamer-full:gst-full-libraries=gstreamer-controller-1.0,gstreamer-plugins-base-1.0,gstreamer-pbutils-1.0" \
		-D "gstreamer-full:gst-full-target-type=static_library" \
		-D "gstreamer-full:gst-full-elements=coreelements:fdsrc,fdsink,queue,queue2,typefind,tee" \
		-D "gstreamer-full:bad=enabled" \
		-D "gstreamer-full:ugly=disabled" \
		-D "gstreamer-full:tls=disabled" \
		-D "gstreamer-full:gpl=enabled" \
		-D "gstreamer-full:gst-full-typefind-functions="

.PHONY: meson-setup
meson-setup:
	meson setup build $(OPTS)
	meson configure build $(OPTS)

.PHONY: node-all-platforms
node-all-platforms: app
	meson setup build-linux-amd64 $(OPTS) --buildtype debugoptimized
	meson compile -C build-linux-amd64 archive
	$(MAKE) link-test
	$(MAKE) linux-arm64
	$(MAKE) windows-amd64
	$(MAKE) windows-amd64-startup-test
	$(MAKE) desktop-linux

.PHONY: desktop-linux
desktop-linux:
	cd js/desktop \
	&& yarn run make --platform win32 --arch x64 \
	&& cd - \
	&& mv "js/desktop/out/make/squirrel.windows/x64/Aquareum-$(subst v,,$(VERSION)) Setup.exe" ./bin/aquareum-desktop-$(VERSION)-windows-amd64.exe

.PHONY: linux-arm64
linux-arm64:
	rustup target add aarch64-unknown-linux-gnu
	meson setup --cross-file util/linux-arm64-gnu.ini --buildtype debugoptimized build-linux-arm64 $(OPTS)
	meson compile -C build-linux-arm64 archive

.PHONY: windows-amd64
windows-amd64:
	rustup target add x86_64-pc-windows-gnu
	meson setup --cross-file util/windows-amd64-gnu.ini --buildtype debugoptimized build-windows-amd64 $(OPTS)
	meson compile -C build-windows-amd64 archive 2>&1 | grep -v drectve

# unbuffer here is a workaround for wine trying to pop up a terminal window and failing
.PHONY: windows-amd64-startup-test
windows-amd64-startup-test:
	bash -c 'set -euo pipefail && unbuffer wine64 ./build-windows-amd64/aquareum.exe --version | cat'

.PHONY: node-all-platforms-macos
node-all-platforms-macos: app
	meson setup --buildtype debugoptimized build-darwin-arm64 $(OPTS)
	meson compile -C build-darwin-arm64 archive
	./build-darwin-arm64/aquareum --version
	rustup target add x86_64-apple-darwin
	meson setup --buildtype debugoptimized --cross-file util/darwin-amd64-apple.ini build-darwin-amd64 $(OPTS)
	meson compile -C build-darwin-amd64 archive
	./build-darwin-amd64/aquareum --version
	$(MAKE) desktop-macos
	meson test -C build-darwin-arm64 go-tests

.PHONY: desktop-macos
desktop-macos:
	cd js/desktop \
	&& yarn run make --platform darwin --arch arm64 \
	&& yarn run make --platform darwin --arch x64 \
	&& cd - \
	&& mv js/desktop/out/make/Aquareum-$(subst v,,$(VERSION))-x64.dmg ./bin/aquareum-desktop-$(VERSION)-darwin-amd64.dmg \
	&& mv js/desktop/out/make/Aquareum-$(subst v,,$(VERSION))-arm64.dmg ./bin/aquareum-desktop-$(VERSION)-darwin-arm64.dmg

# link your local version of mist for dev
.PHONY: link-mist
link-mist:
	rm -rf subprojects/mistserver
	ln -s $$(realpath ../mistserver) ./subprojects/mistserver

# link your local version of c2pa-go for dev
.PHONY: link-c2pa-go
link-c2pa-go:
	rm -rf subprojects/c2pa_go
	ln -s $$(realpath ../c2pa-go) ./subprojects/c2pa_go

# link your local version of gstreamer
.PHONY: link-gstreamer
link-gstreamer:
	rm -rf subprojects/gstreamer-full
	ln -s $$(realpath ../gstreamer) ./subprojects/gstreamer-full

# link your local version of ffmpeg for dev
.PHONY: link-ffmpeg
link-ffmpeg:
	rm -rf subprojects/FFmpeg
	ln -s $$(realpath ../ffmpeg) ./subprojects/FFmpeg

.PHONY: docker-build
docker-build: docker-build-builder docker-build-in-container

.PHONY: docker-build-builder
docker-build-builder:
	cd docker \
	&& docker build --target=builder --os=linux --arch=amd64 -f build.Dockerfile -t aqrm.io/aquareum-tv/aquareum:builder .

.PHONY: docker-build-builder
docker-build-in-container:
	docker run -v $$(pwd):$$(pwd) -w $$(pwd) --rm -it aqrm.io/aquareum-tv/aquareum:builder make app-and-node

.PHONY: docker-release
docker-release:
	cd docker \
	&& docker build -f release.Dockerfile \
	  --build-arg TARGETARCH=$(BUILDARCH) \
		-t aqrm.io/aquareum-tv/aquareum \
		.

.PHONY: ci-upload
ci-upload: ci-upload-node ci-upload-android

.PHONY: ci-upload-node
ci-upload-node: node-all-platforms
	for GOOS in linux; do \
		for GOARCH in amd64 arm64; do \
			export file=aquareum-$(VERSION)-$$GOOS-$$GOARCH.tar.gz \
			&& $(MAKE) ci-upload-file upload_file=$$file; \
		done \
	done;
	for GOOS in windows; do \
		for GOARCH in amd64; do \
			export file=aquareum-$(VERSION)-$$GOOS-$$GOARCH.zip \
			&& $(MAKE) ci-upload-file upload_file=$$file; \
			export file=aquareum-desktop-$(VERSION)-$$GOOS-$$GOARCH.exe \
			&& $(MAKE) ci-upload-file upload_file=$$file; \
		done \
	done;

.PHONY: ci-upload-node-macos
ci-upload-node-macos: node-all-platforms-macos
	for GOOS in darwin; do \
		for GOARCH in amd64 arm64; do \
			export file=aquareum-$(VERSION)-$$GOOS-$$GOARCH.tar.gz \
			&& $(MAKE) ci-upload-file upload_file=$$file; \
			export file=aquareum-desktop-$(VERSION)-$$GOOS-$$GOARCH.dmg \
			&& $(MAKE) ci-upload-file upload_file=$$file; \
		done \
	done;

.PHONY: ci-upload-android
ci-upload-android: android
	$(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-android-release.apk \
	&& $(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-android-debug.apk \
	&& $(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-android-release.aab \
	&& $(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-android-debug.aab

.PHONY: ci-upload-ios
ci-upload-ios: ios
	$(MAKE) ci-upload-file upload_file=aquareum-$(VERSION)-ios-release.xcarchive.tar.gz

upload_file?=""
.PHONY: ci-upload-file
ci-upload-file:
	curl \
		--fail-with-body \
		--header "JOB-TOKEN: $$CI_JOB_TOKEN" \
		--upload-file bin/$(upload_file) \
		"$$CI_API_V4_URL/projects/$$CI_PROJECT_ID/packages/generic/$(BRANCH)/$(VERSION)/$(upload_file)";

.PHONY: release
release:
	yarn run release

.PHONY: ci-release
ci-release:
	go install gitlab.com/gitlab-org/release-cli/cmd/release-cli
	go run ./pkg/config/git/git.go -release -o release.yml
	release-cli create-from-file --file release.yml

.PHONY: check
check: install
	yarn run check

.PHONY: fix
fix:
	yarn run fix

.PHONY: precommit
precommit: dockerfile-hash-precommit

.PHONY: dockefile-hash-precommit
dockerfile-hash-precommit:
	@bash -c 'printf "variables:\n  DOCKERFILE_HASH: `git hash-object docker/build.Dockerfile`" > .ci/dockerfile-hash.yaml' \
	&& git add .ci/dockerfile-hash.yaml
