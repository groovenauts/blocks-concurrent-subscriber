PKGDIR=./pkg
BASENAME=blocks-concurrent-subscriber
VERSION=`grep VERSION version.go | cut -f2 -d\"`
OS=linux
ARCH=amd64

all: build

setup:
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr

checksetup:
	go get golang.org/x/tools/cmd/goimports

testsetup:
	mysql -u root -e "CREATE DATABASE IF NOT EXISTS blocks_subscriber_test DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;"
	mysql -u root blocks_subscriber_test < migrations/up.sql

testfixtures:
	mysql -u root blocks_subscriber_test < test/setup.sql

testteardown:
	mysql -u root blocks_subscriber_test < migrations/down.sql
	mysql -u root -e "DROP DATABASE IF EXISTS blocks_subscriber_test;"

check: checkfmt
	go vet *.go
	goimports -l *.go

checkfmt:
ifneq ($(shell gofmt -l *.go),)
	exit 1
else
	@echo "gofmt -l *.go OK"
endif

build:
	mkdir -p ${PKGDIR}
	gox -output="${PKGDIR}/{{.Dir}}_${OS}_${ARCH}" -os="${OS}" -arch="${ARCH}"

release: build
	ghr -u groovenauts -r blocks-concurrent-subscriber --replace --draft ${VERSION} pkg

prerelease: build
	ghr -u groovenauts -r blocks-concurrent-subscriber --replace --draft --prerelease ${VERSION} pkg

version:
	echo ${VERSION}

clean:
	rm -rf ${PKGDIR}
