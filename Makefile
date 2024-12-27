# 当前 HEAD 的 tag
HEAD_TAG = $(shell git describe --exact-match --tags 2>/dev/null)

# 当前分支名
HEAD_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

# 当前 commit hash
HEAD_HASH = $(shell git rev-parse HEAD)

# 1. 如果HEAD存在tag，则GIT_VERSION=<版本名称>-<企业版/社区版> <commit>
# PS: 通常会在版本名称前增加字符“v”作为tag内容，当版本名称为 3.2411.0时，tag内容为v3.2411.0 
# e.g. tag为v3.2411.0时，社区版：GIT_VERSION=3.2411.0-ce a6355ff4cf8d181315a2b30341bc954b29576b11
# e.g. tag为v3.2412.0-pre1-1时，社区版：GIT_VERSION=3.2412.0-pre1-1-ce f0bcb90e712cbdb6e16f122c1ebd623e90f9a905
# 2. 如果HEAD没有tag，则GIT_VERSION=<分支名> <commit>
# e.g. 分支名为main时，GIT_VERSION=main a6355ff4cf8d181315a2b30341bc954b29576b11
# e.g. 分支名为release-3.2411.x时，GIT_VERSION=release-3.2411.x a6355ff4cf8d181315a2b30341bc954b29576b11
override GIT_VERSION = $(if $(HEAD_TAG), \
    $(shell echo $(HEAD_TAG) | sed 's/^v//')-$(EDITION), \
    $(HEAD_BRANCH))${CUSTOM} $(HEAD_HASH)
override GIT_COMMIT     		= $(shell git rev-parse HEAD)
override PROJECT_NAME 			= dms
override DOCKER         		= $(shell which docker)
override GOOS           		= linux
override OS_VERSION 			= el7
override GO_BUILD_FLAGS 		= -mod=vendor
override RPM_USER_GROUP_NAME 	= actiontech
override RPM_USER_NAME 			= actiontech-universe
override LDFLAGS 				= -ldflags "-X 'main.Version=${GIT_VERSION}' -X 'main.gitCommitID=${GIT_COMMIT}' -X 'main.defaultRunUser=${RPM_USER_NAME}'"

GO_COMPILER_IMAGE ?= golang:1.19.6
RPM_BUILD_IMAGE ?= rpmbuild/centos7
DOCKER_IMAGE_GO_SWAGGER = quay.io/goswagger/swagger
DMS_UNIT_TEST_MYSQL_DB_CONTAINER = dms-mysql-unit-test-db
ARCH=$(shell uname -s| tr A-Z a-z)

GOARCH         		= amd64
RPMBUILD_TARGET		= x86_64
ifeq ($(GOARCH), arm64)
    RPMBUILD_TARGET = aarch64
endif

EDITION ?= ce
GO_BUILD_TAGS = dummyhead
ifeq ($(EDITION),ee)
    GO_BUILD_TAGS :=$(GO_BUILD_TAGS),enterprise
else ifeq ($(EDITION),trial)
    GO_BUILD_TAGS :=$(GO_BUILD_TAGS),trial
endif
RELEASE = qa
ifeq ($(RELEASE),rel)
    GO_BUILD_TAGS :=$(GO_BUILD_TAGS),release
endif

PRODUCT_CATEGORY =
ifeq ($(PRODUCT_CATEGORY),dms)
    GO_BUILD_TAGS :=$(GO_BUILD_TAGS),dms
endif

ifeq ($(IS_PRODUCTION_RELEASE),true)
# When performing a publishing operation, two cases:
# 1. if there is tag on current commit, means that
# 	 we release new version on current branch just now.
#    Set rpm name with tag name(v1.2109.0 -> 1.2109.0).
#
# 2. if there is no tag on current commit, means that
#    current branch is on process.
#    Set rpm name with current branch name(release-1.2109.x-ee or release-1.2109.x -> 1.2109.x).
    PROJECT_VERSION = $(shell if [ "$$(git tag --points-at HEAD | tail -n1)" ]; then git tag --points-at HEAD | tail -n1 | sed 's/v\(.*\)/\1/'; else git rev-parse --abbrev-ref HEAD | sed 's/release-\(.*\)/\1/' | tr '-' '\n' | head -n1; fi)
else
#    When performing daily packaging, set rpm name with current branch name(release-1.2109.x-ee or release-1.2109.x -> 1.2109.x).
    PROJECT_VERSION = $(shell git rev-parse --abbrev-ref HEAD | sed 's/release-\(.*\)/\1/' | tr '-' '\n' | head -n1)
endif

override RPM_NAME = $(PROJECT_NAME)-$(EDITION)-$(PROJECT_VERSION).$(RELEASE).$(OS_VERSION).$(RPMBUILD_TARGET).rpm

ADDITIONAL_PROJECT_NAME ?=
ifdef ADDITIONAL_PROJECT_NAME
	override RPM_NAME := $(PROJECT_NAME)-$(ADDITIONAL_PROJECT_NAME)-$(EDITION)-$(PROJECT_VERSION).$(RELEASE).$(OS_VERSION).$(RPMBUILD_TARGET).rpm
endif

override FTP_PATH = ftp://$(RELEASE_FTPD_HOST)/actiontech-$(PROJECT_NAME)/$(EDITION)/$(RELEASE)/$(PROJECT_VERSION)/$(RPM_NAME)

############################### compiler ##################################
dlv_install:
	GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.io,direct go build -gcflags "all=-N -l" $(GO_BUILD_FLAGS) ${LDFLAGS} -tags $(GO_BUILD_TAGS) -o ./bin/dms ./internal/apiserver/cmd/server/main.go

install:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_BUILD_FLAGS) ${LDFLAGS} -tags $(GO_BUILD_TAGS) -o ./bin/dms ./internal/apiserver/cmd/server/main.go

docker_install:
	$(DOCKER) run -v $(shell pwd):/universe --rm $(GO_COMPILER_IMAGE) sh -c "cd /universe && git config --global --add safe.directory /universe && make install $(MAKEFLAGS)"

docker_rpm: docker_install
	$(DOCKER) run -v $(shell pwd):/universe/dms --user root --rm -e VERBOSE=1 $(RPM_BUILD_IMAGE) sh -c "(mkdir -p /root/rpmbuild/SOURCES >/dev/null 2>&1);cd /root/rpmbuild/SOURCES; \
	(tar zcf ${PROJECT_NAME}.tar.gz /universe --transform 's/universe/${PROJECT_NAME}-$(GIT_COMMIT)/' >/tmp/build.log 2>&1) && \
	(rpmbuild --define 'group_name $(RPM_USER_GROUP_NAME)' --define 'user_name $(RPM_USER_NAME)' \
	--define 'commit $(GIT_COMMIT)' --define 'os_version $(OS_VERSION)' --define 'project_name $(PROJECT_NAME)' \
	--target $(RPMBUILD_TARGET)  -bb -vv --with qa /universe/dms/build/dms.spec >>/tmp/build.log 2>&1) && \
	(cat /root/rpmbuild/RPMS/$(RPMBUILD_TARGET)/${PROJECT_NAME}-$(GIT_COMMIT)-$(OS_VERSION).$(RPMBUILD_TARGET).rpm) || (cat /tmp/build.log && exit 1)" > $(RPM_NAME) && \
	md5sum $(RPM_NAME) > $(RPM_NAME).md5

override PRE_DIR = $(dir $(CURDIR))

dms_sqle_provision_rpm_pre: docker_install
	rm -rf builddir

	mkdir -p ./builddir/bin
	mkdir -p ./builddir/config
	mkdir -p ./builddir/static/logo
	mkdir -p ./builddir/scripts
	mkdir -p ./builddir/neo4j-community
	mkdir -p ./builddir/lib

	# 前端文件
	cp -R ${PRE_DIR}dms-ui/packages/base/dist/* ./builddir/static/

	# dms文件
	cp ./bin/dms ./builddir/bin/dms
	cp ./config.yaml ./builddir/config/dms.yaml
	cp ./build/service-file-template/dms.systemd ./builddir/scripts/dms.systemd
	cp ./build/scripts/dms_sqle_provision.sh ./builddir/scripts/init_start.sh
	cp ./build/logo/* ./builddir/static/logo/

	# provision文件
	cp ${PRE_DIR}provision/bin/provision ./builddir/bin/provision
	cp ${PRE_DIR}provision/build/service-file-template/neo4j.systemd ./builddir/scripts/neo4j.systemd
	cp ${PRE_DIR}provision/build/service-file-template/provision.systemd ./builddir/scripts/provision.systemd
	cp ${PRE_DIR}provision/config.yaml ./builddir/config/provision.yaml
	cp -R ${PRE_DIR}provision/build/neo4j-community/* ./builddir/neo4j-community/
	cp -R ${PRE_DIR}provision/lib/* ./builddir/lib/

	# sqle文件
	cp ${PRE_DIR}sqle/config.yaml.template ./builddir/config/sqle.yaml
	cp ${PRE_DIR}sqle/bin/sqled ./builddir/bin/sqled
	cp ${PRE_DIR}sqle/bin/scannerd ./builddir/bin/scannerd
	cp ${PRE_DIR}sqle/scripts/sqled.systemd ./builddir/scripts/sqled.systemd

	# 合并配置文件
	touch ./builddir/config/config.yaml
	cat ./builddir/config/dms.yaml >> ./builddir/config/config.yaml
	cat ./builddir/config/sqle.yaml >> ./builddir/config/config.yaml
	cat ./builddir/config/provision.yaml >> ./builddir/config/config.yaml
	rm ./builddir/config/dms.yaml ./builddir/config/sqle.yaml ./builddir/config/provision.yaml

# 本地打包需要先编译sqle,provision,dms-ui相关文件
dms_sqle_provision_rpm: dms_sqle_provision_rpm_pre
	$(DOCKER) run -v $(shell pwd):/universe/dms --user root --rm $(RPM_BUILD_IMAGE) sh -c "(mkdir -p /root/rpmbuild/SOURCES >/dev/null 2>&1);cd /root/rpmbuild/SOURCES; \
	(tar zcf ${PROJECT_NAME}.tar.gz /universe/dms --transform 's/universe/${PROJECT_NAME}-$(GIT_COMMIT)/' > /tmp/build.log 2>&1) && \
	(rpmbuild --define 'group_name $(RPM_USER_GROUP_NAME)' --define 'user_name $(RPM_USER_NAME)' --define 'commit $(GIT_COMMIT)' --define 'os_version $(OS_VERSION)' \
	--target $(RPMBUILD_TARGET) --with qa -bb /universe/dms/build/dms_sqle_provision.spec >> /tmp/build.log 2>&1) && \
	(cat ~/rpmbuild/RPMS/$(RPMBUILD_TARGET)/${PROJECT_NAME}-$(GIT_COMMIT)-qa.$(OS_VERSION).$(RPMBUILD_TARGET).rpm) || (cat /tmp/build.log && exit 1)" > $(RPM_NAME) && \
	md5sum $(RPM_NAME) > $(RPM_NAME).md5

upload_rpm:
	curl -T $(shell pwd)/$(RPM_NAME) $(FTP_PATH) --ftp-create-dirs
	curl -T $(shell pwd)/$(RPM_NAME).md5 $(FTP_PATH).md5 --ftp-create-dirs

############################### check ##################################
validation_swag:
	$(DOCKER) run --rm -e VERBOSE=1 --volume=$(shell pwd):/source $(DOCKER_IMAGE_GO_SWAGGER) validate /source/api/swagger.json

golangci_lint:
	golangci-lint run -c ./build/golangci-lint/.golangci.yml --timeout=10m
# check go lint on dev
golangci_lint_dev:
	docker run --rm -v $(shell pwd):/src  golangci/golangci-lint:v1.49 bash -c "cd /src && make golangci_lint"

############################### test ##################################
dms_unit_test_prepare:
	./build/scripts/dms_run_unit_test_db_container.sh $(DMS_UNIT_TEST_MYSQL_DB_CONTAINER)

dms_unit_test_clean:
	docker rm -f $(DMS_UNIT_TEST_MYSQL_DB_CONTAINER)

dms_test_dms:
	go test -v -p 1 ./internal/dms/...

############################### generate ##################################
gen_repo_fields:
	go run ./internal/dms/cmd/gencli/gencli.go -d generate-node-repo-fields ./internal/dms/storage/model/ ./internal/dms/biz/

gen_swag:
	./internal/apiserver/cmd/swag/swagger_${ARCH}_amd64 generate spec -m -w ./internal/apiserver/cmd/server/ -o ./api/swagger.yaml
	./internal/apiserver/cmd/swag/swagger_${ARCH}_amd64 generate spec -i ./api/swagger.yaml -o ./api/swagger.json

open_swag_server:
	./internal/apiserver/cmd/swag/swagger_${ARCH}_amd64 serve --no-open -F=swagger --port 36666 ./api/swagger.yaml

######################################## i18n ##########################################################
GOBIN = ${shell pwd}/bin
LOCALE_PATH   = ${shell pwd}/internal/pkg/locale

install_i18n_tool:
	GOBIN=$(GOBIN) go install -v github.com/nicksnyder/go-i18n/v2/goi18n@latest

extract_i18n:
	cd ${LOCALE_PATH} && $(GOBIN)/goi18n extract -sourceLanguage zh

start_trans_i18n:
	cd ${LOCALE_PATH} && touch translate.en.toml && $(GOBIN)/goi18n merge -sourceLanguage=zh active.*.toml

end_trans_i18n:
	cd ${LOCALE_PATH} && $(GOBIN)/goi18n merge active.en.toml translate.en.toml && rm -rf translate.en.toml
