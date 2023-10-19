override GIT_VERSION    		= $(shell git rev-parse --abbrev-ref HEAD)${CUSTOM} $(shell git rev-parse HEAD)
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
endif

RELEASE = qa
ifeq ($(RELEASE),rel)
    GO_BUILD_TAGS :=$(GO_BUILD_TAGS),release
endif

# Two cases:
# 1. if there is tag on current commit, means that
# 	 we release new version on current branch just now.
#    Set rpm name with tag name(v1.2109.0 -> 1.2109.0).
#
# 2. if there is no tag on current commit, means that
#    current branch is on process.
#    Set rpm name with current branch name(release-1.2109.x-ee or release-1.2109.x -> 1.2109.x).
PROJECT_VERSION = $(shell if [ "$$(git tag --points-at HEAD | tail -n1)" ]; then git tag --points-at HEAD | tail -n1 | sed 's/v\(.*\)/\1/'; else git rev-parse --abbrev-ref HEAD | sed 's/release-\(.*\)/\1/' | tr '-' '\n' | head -n1; fi)

override RPM_NAME = $(PROJECT_NAME)-$(EDITION)-$(PROJECT_VERSION).$(RELEASE).$(OS_VERSION).$(RPMBUILD_TARGET).rpm

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