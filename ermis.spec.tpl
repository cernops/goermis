%global provider	gitlab
%global provider_tld	cern.ch
%global project		lb-experts
%global provider_full %{provider}.%{provider_tld}/%{project}
%global repo		goermis

%global import_path	%{provider_full}/%{repo}
%global gopath		%{_datadir}/gocode
%global debug_package	%{nil}

Name: ermis
Version: #REPLACE_BY_VERSION#
Release: #REPLACE_BY_RELEASE#%{?dist}

Summary: CERN LB DNS Web interface
License: ASL 2.0
URL: https://%{import_path}
Source: %{name}-%{version}.tgz
BuildRequires: golang >= 1.14
ExclusiveArch: x86_64

%description
%{summary}

Web interface for ermis

%prep
%setup -n %{name}-%{version} -q

%build
go build -o ermis -mod=vendor


%install
# main package binary
install -d -p %{buildroot}/usr/sbin/ %{buildroot}/var/lib/ermis/ %{buildroot}/lib/systemd/system/
install -p -m0755 ermis %{buildroot}/usr/sbin/ermis
cp -r staticfiles templates %{buildroot}/var/lib/ermis/
install -p -m0644 config/systemd/ermis.service  %{buildroot}/lib/systemd/system

%files
%doc LICENSE COPYING README.md
/usr/sbin/ermis
/var/lib/ermis/
/lib/systemd/system/ermis.service

%changelog
* Mon Jul 05 2021 Kristian Kouros <kristian.kouros@cern.ch> - 1.3.0-5
- build for lb8s
* Mon Jul 05 2021 Kristian Kouros <kristian.kouros@cern.ch> - 1.3.0-4
- edit ermis.service
* Fri Jul 03 2020 Pablo Saiz <pablo.saiz@cern.ch>           - 0.0.3
- Include staticfiles
* Mon May 25 2020 Pablo Saiz <pablo.saiz@cern.ch>           - 0.0.2
- Include the service startup
* Wed May 20 2020 Pablo Saiz <pablo.saiz@cern.ch>           - 0.0.1
- First version
