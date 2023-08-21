PROJECT_NAME              = dms
VERSION                   = 99.99.99
RELEASE                   ?= alpha
USER_NAME                 = actiontech-$(PROJECT_NAME)
GROUP_NAME                = actiontech
GIT_LATEST_COMMIT_ID	  = $(shell git rev-parse HEAD)
DMS_UNIT_TEST_MYSQL_DB_CONTAINER = dms-mysql-unit-test-db
ARCH=$(shell uname -s| tr A-Z a-z)
DOCKER=$(shell which docker)
DOCKER_IMAGE_RPM = reg.actiontech.com/actiontech/$(PROJECT_NAME)-rpmbuild:v1
DOCKER_IMAGE_GO_SWAGGER = quay.io/goswagger/swagger
GIT_LATEST_COMMIT_ID	  = $(shell git rev-parse HEAD)
EDITION ?= ce
GO_BUILD_TAGS = dummyhead
ifeq ($(EDITION),ee)
    GO_BUILD_TAGS :=$(GO_BUILD_TAGS),enterprise
endif
RPM_NAME            = $(PROJECT_NAME)-$(EDITION)-$(VERSION)-${RELEASE}.el7.x86_64.rpm
RPM_BUILD_PATH      = ./RPMS/x86_64/$(RPM_NAME)

gen_repo_fields:
	go run ./internal/dms/cmd/gencli/gencli.go -d generate-node-repo-fields ./internal/dms/storage/model/ ./internal/dms/biz/
build_dms:
	GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.io,direct go build -tags ${GO_BUILD_TAGS} -mod=vendor -ldflags "-X 'main.defaultRunUser=${USER_NAME}' -X 'main.version=${VERSION}' -X 'main.gitCommitID=${GIT_LATEST_COMMIT_ID}'" -o ./bin/dms ./internal/apiserver/cmd/server/main.go
dms_unit_test_prepare:
	./build/scripts/dms_run_unit_test_db_container.sh $(DMS_UNIT_TEST_MYSQL_DB_CONTAINER)
dms_unit_test_clean:
	docker rm -f $(DMS_UNIT_TEST_MYSQL_DB_CONTAINER)
dms_test_dms:
	go test -v -p 1 ./internal/dms/...
gen_swag:
	./internal/apiserver/cmd/swag/swagger_${ARCH}_amd64 generate spec -m -w ./internal/apiserver/cmd/server/ -o ./api/swagger.yaml
	./internal/apiserver/cmd/swag/swagger_${ARCH}_amd64 generate spec -i ./api/swagger.yaml -o ./api/swagger.json
validation_swag:
	$(DOCKER) run --rm -e VERBOSE=1 --volume=$(shell pwd):/source $(DOCKER_IMAGE_GO_SWAGGER) validate /source/api/swagger.json
open_swag_server:
	./internal/apiserver/cmd/swag/swagger_${ARCH}_amd64 serve --no-open -F=swagger --port 36666 ./api/swagger.yaml
upload_rpm:
	curl -T $(RPM_BUILD_PATH) ftp://$(RELEASE_FTPD_HOST)/actiontech-$(PROJECT_NAME)/$(EDITION)/$(VERSION)/$(RPM_NAME) --ftp-create-dirs

rpm: build_dms
	$(DOCKER) run --rm -e VERBOSE=1 --volume=$(shell pwd):/source --volume=$(shell pwd)/build:/spec --volume=$(shell pwd):/out \
	--env=PRE_BUILDDEP="mkdir /src && tar zcf  /src/$(PROJECT_NAME).tar.gz /source --transform 's|source|$(PROJECT_NAME)-$(VERSION)|'" \
	--env=RPM_ARGS='-bb \
	--define "user_name $(USER_NAME)" --define "group_name $(GROUP_NAME)" --define "project_name $(PROJECT_NAME)" --define "version $(VERSION)" --define "release $(RELEASE)"' \
	$(DOCKER_IMAGE_RPM) dms.spec &&\
	mv ./RPMS/x86_64/$(PROJECT_NAME)-$(VERSION)-${RELEASE}.el7.x86_64.rpm ./RPMS/x86_64/$(RPM_NAME)
	
download_front:
	wget ftp://$(RELEASE_FTPD_HOST)/actiontech-dms-ui/$(VERSION)/dms-ui-$(VERSION).tar.gz -O ./build/dms-ui-$(VERSION).tar.gz
	mkdir -p ./build/static
	tar zxf ./build/dms-ui-$(VERSION).tar.gz --strip-components 4 -C ./build/static/

golangci_lint:
	golangci-lint run -c ./build/golangci-lint/.golangci.yml --timeout=10m
# check go lint on dev
golangci_lint_dev:
	docker run --rm -v $(shell pwd):/src  golangci/golangci-lint:v1.49 bash -c "cd /src && make golangci_lint"
dlv_install:
	GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.io,direct go build -gcflags "all=-N -l" -tags ${RELEASE} -mod=vendor -ldflags "-X 'main.defaultRunUser=${USER_NAME}' -X 'main.version=${VERSION}' -X 'main.gitCommitID=${GIT_LATEST_COMMIT_ID}'" -o ./bin/dms ./internal/apiserver/cmd/server/main.go