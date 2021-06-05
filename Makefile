VERSION         ?= 0.0.1

PACK            := miniflux
PROJECT         := github.com/cnunciato/pulumi-${PACK}

PROVIDER        := pulumi-resource-${PACK}
CODEGEN         := pulumi-gen-${PACK}
VERSION_PATH    := provider/pkg/version.Version

WORKING_DIR     := $(shell pwd)
SCHEMA_PATH     := ${WORKING_DIR}/schema.json

PROVIDER_BINARY	:= ${PROVIDER}-v${VERSION}-darwin-amd64.tar.gz

generate:: gen_go_sdk gen_dotnet_sdk gen_nodejs_sdk gen_python_sdk

build:: build_provider build_dotnet_sdk build_nodejs_sdk build_python_sdk

install:: install_provider install_dotnet_sdk install_nodejs_sdk

publish:: publish_provider publish_nodejs_sdk publish_dotnet_sdk publish_python_sdk

# Provider

build_provider::
	rm -rf ${WORKING_DIR}/bin/${PROVIDER}
	VERSION=${VERSION} && sed -i.bak -e "s/[0-9]*\.[0-9]*\.[0-9]*/${VERSION}/" provider/pkg/version/version.go
	cd provider/cmd/${PROVIDER} && VERSION=${VERSION} SCHEMA=${SCHEMA_PATH} go generate main.go
	cd provider/cmd/${PROVIDER} && go build -a -o ${WORKING_DIR}/bin/${PROVIDER} -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" .

install_provider:: build_provider
	cp ${WORKING_DIR}/bin/${PROVIDER} ${GOPATH}/bin

publish_provider::
	cd bin && tar -czvf ${PROVIDER_BINARY} ${PROVIDER} && cd ..
	aws s3 cp ./bin/${PROVIDER_BINARY} s3://cnunciato-pulumi-components --acl public-read --region us-west-2


# Go SDK

gen_go_sdk::
	rm -rf sdk/go
	cd provider/cmd/${CODEGEN} && go run . go ../../../sdk/go ${SCHEMA_PATH}


# .NET SDK

gen_dotnet_sdk::
	rm -rf sdk/dotnet
	cd provider/cmd/${CODEGEN} && go run . dotnet ../../../sdk/dotnet ${SCHEMA_PATH}

build_dotnet_sdk:: DOTNET_VERSION := ${VERSION}
build_dotnet_sdk:: gen_dotnet_sdk
	cd sdk/dotnet/ && \
		echo "${DOTNET_VERSION}" >version.txt && \
		dotnet build /p:Version=${DOTNET_VERSION}

install_dotnet_sdk:: build_dotnet_sdk
	rm -rf ${WORKING_DIR}/nuget
	mkdir -p ${WORKING_DIR}/nuget
	find . -name '*.nupkg' -print -exec cp -p {} ${WORKING_DIR}/nuget \;


# Node.js SDK

gen_nodejs_sdk::
	rm -rf sdk/nodejs
	cd provider/cmd/${CODEGEN} && go run . nodejs ../../../sdk/nodejs ${SCHEMA_PATH}

build_nodejs_sdk:: gen_nodejs_sdk
	cd sdk/nodejs/ && \
		yarn install && \
		yarn run tsc --version && \
		yarn run tsc && \
		cp ../../doc/nodejs/README.md ../../LICENSE package.json yarn.lock ./bin/ && \
		sed -i.bak -e "s/\$${VERSION}/$(VERSION)/g" ./bin/package.json && \
		sed -i.bak -e "s/\$${VERSION}/$(VERSION)/g" ./bin/README.md && \
		rm ./bin/package.json.bak

install_nodejs_sdk:: build_nodejs_sdk
	yarn link --cwd ${WORKING_DIR}/sdk/nodejs/bin

publish_nodejs_sdk::
	cd sdk/nodejs/bin && npm publish

publish_dotnet_sdk::
	cd nuget && dotnet nuget push Pulumi.Miniflux.${VERSION}.nupkg --api-key ${NUGET_API_KEY} --source https://api.nuget.org/v3/index.json

publish_python_sdk::
	cd sdk/python && python3 setup.py sdist bdist_wheel && twine upload dist/*

# Python SDK

gen_python_sdk::
	rm -rf sdk/python
	cd provider/cmd/${CODEGEN} && go run . python ../../../sdk/python ${SCHEMA_PATH}
	cp ${WORKING_DIR}/README.md sdk/python

build_python_sdk:: PYPI_VERSION := ${VERSION}
build_python_sdk:: gen_python_sdk
	cd sdk/python/ && \
		python3 setup.py clean --all 2>/dev/null && \
		rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
		sed -i.bak -e "s/\$${VERSION}/${PYPI_VERSION}/g" -e "s/\$${PLUGIN_VERSION}/${VERSION}/g" ./bin/setup.py && \
		rm ./bin/setup.py.bak && \
		cd ./bin && python3 setup.py build sdist
