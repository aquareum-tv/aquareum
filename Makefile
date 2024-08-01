OUT_DIR?="bin"
$(shell mkdir -p $(OUT_DIR))

.PHONY: default
default: app node

VERSION?=$(shell go run ./pkg/config/git/git.go -v)
UUID?=$(shell go run ./pkg/config/uuid/uuid.go)
BRANCH?=$(shell go run ./pkg/config/git/git.go --branch)

.PHONY: version
version:
	@go run ./pkg/config/git/git.go -v \
	&& go run ./pkg/config/git/git.go --env -o .ci/build.env

.PHONY: install
install:
	yarn install --inline-builds

.PHONY: app
app: install
	yarn run build

.PHONY: node
node:
	meson setup build --native=./util/linux-amd64-gnu.ini && meson compile -C build
	mv ./build/aquareum ./bin/aquareum

.PHONY: test
test: app
	go test ./pkg/... ./cmd/...

.PHONY: all
all: version install check app test node-all-platforms android

.PHONY: ci
ci: version install check app test node-all-platforms ci-upload-node

.PHONY: ci-macos
ci-macos: version install check app ios ci-upload-ios

.PHONY: ci-macos
ci-android: version install check android ci-upload-android

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
		clean archive \
	&& cd bin \
	&& tar -czvf aquareum-$(VERSION)-ios-release.xcarchive.tar.gz aquareum-$(VERSION)-ios-release.xcarchive

# xcodebuild -exportArchive -archivePath ./bin/aquareum-$(VERSION)-ios-release.xcarchive -exportOptionsPlist ./js/app/exportOptions.plist -exportPath ./bin/aquareum-$(VERSION)-ios-release.ipa

.build/bundletool.jar:
	mkdir -p .build \
	&& curl -L -o ./.build/bundletool.jar https://github.com/google/bundletool/releases/download/1.17.0/bundletool-all-1.17.0.jar

.PHONY: node-all-platforms
node-all-platforms: app
	meson setup build
	meson compile -C build archive
	meson setup --cross-file util/linux-arm64-gnu.ini build-aarch64
	meson compile -C build-aarch64 archive

# link your local version of mist for dev
.PHONY: link-mist
link-mist:
	rm -rf subprojects/mistserver
	ln -s $$(realpath ../mistserver) ./subprojects/mistserver

.PHONY: docker-build
docker-build: docker-build-builder docker-build-in-container

.PHONY: docker-build-builder
docker-build-builder:
	cd docker \
	&& docker build --target=builder --os=linux --arch=amd64 -f build.Dockerfile -t aqrm.io/aquareum-tv/aquareum:builder .

.PHONY: docker-build-builder
docker-build-in-container:
	docker run -v $$(pwd):$$(pwd) -w $$(pwd) --rm -it aqrm.io/aquareum-tv/aquareum:builder make

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
	curl --silent --fail "$$CI_API_V4_URL/projects/$$CI_PROJECT_ID/repository/changelog?version=$(VERSION)" | jq -r '.notes' > description.md
	release-cli create \
		--name $(VERSION) \
		--tag-name $(VERSION) \
		--description description.md \
		--assets-link '$(shell ./util/release-files.sh $(VERSION))'

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
