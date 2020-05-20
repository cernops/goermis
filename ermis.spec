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

Summary: CERN DNS ERMIS Web interface
License: ASL 2.0
URL: https://%{import_path}
Source: %{name}-%{version}.tgz
BuildRequires: golang >= 1.5
BuildRequires: checkpolicy
%if 0%{?el6}%{?el7}
BuildRequires: policycoreutils-python
%else
BuildRequires: policycoreutils-python-utils
%endif
ExclusiveArch: x86_64
Requires: net-snmp

%description
%{summary}

Web interface for ermis

%prep
%setup -n %{name}-%{version} -q

%build
mkdir -p src/%{provider_full}
ln -s ../../../ src/%{provider_full}/%{repo}
GOPATH=$(pwd):%{gopath} go build -o ermis %{import_path}


%install
# main package binary
install -d -p %{buildroot}/usr/sbin/ %{buildroot}/usr/share/selinux/targeted/ %{buildroot}/usr/local/etc %{buildroot}/usr/local/sbin/
install -p -m0755 ermis %{buildroot}/usr/sbin/ermis


%files
%doc LICENSE COPYING README.md
/usr/sbin/ermis


%changelog
* Wed May 20 2020 Pablo Saiz <pablo.saiz@cern.ch>           - 0.0.1
- First version
