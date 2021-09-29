DIST ?= $(shell rpm --eval %{dist})
SPECFILE ?= ermis.spec

PKG ?= $(shell rpm -q --specfile $(SPECFILE) --queryformat "%{name}-%{version}\n" | head -n 1)

srpm:
	echo "Creating the source rpm"
	rm -rf build
	mkdir -p SOURCES version build
	go mod edit -replace gitlab.cern.ch/lb-experts/goermis=/builddir/build/BUILD/$(PKG)
	go mod vendor
	tar zcf SOURCES/$(PKG).tgz --exclude build --exclude SOURCES --exclude .git --exclude .koji --exclude .gitlab-ci.yml --transform "s||$(PKG)/|" .
	rpmbuild -bs --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE)

rpm: srpm
	echo "Creating the rpm"
	rpmbuild -bb --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE)
