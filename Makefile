DIST ?= $(shell rpm --eval %{dist})
SPECFILE ?= ermis.spec

PKG ?= $(shell rpm -q --specfile $(SPECFILE) --queryformat "%{name}-%{version}\n" | head -n 1)

installgo:
	mkdir -p /go14
	yum -y install git gcc
	curl https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz  | tar -zxC /go14
	rm -f /usr/bin/go
	ln -s /go14/go/bin/go /usr/bin/go
	export GOPATH=/go14
	go get ./...

srpm: installgo
	echo "Creating the source rpm"
	mkdir -p SOURCES version
	go mod vendor
	tar zcf SOURCES/$(PKG).tgz  --exclude SOURCES --exclude .git --exclude .koji --exclude .gitlab-ci.yml --exclude go.mod --exclude go.sum --transform "s||$(PKG)/|" .
	rpmbuild -bs --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE)

rpm: srpm
	echo "Creating the rpm"
	rpmbuild -bb --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE)
