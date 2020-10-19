FROM cern/cc8-base
WORKDIR /root/
LABEL maintainer="Kristian Kouros <kristian.kouros@cern.ch>"
RUN  mkdir -p  /var/lib/ermis/ /var/log/ermis/
RUN yum-config-manager --add-repo  http://linuxsoft.cern.ch/internal/repos/lb8-stable/x86_64/os  && \ 
     ( yum install -y ermis || true) 
EXPOSE 8080
ENTRYPOINT ["ermis"]
