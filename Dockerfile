FROM cern/cc7-base
WORKDIR /root/
LABEL maintainer="Kristian Kouros <kristian.kouros@cern.ch>"
RUN  mkdir -p  /var/lib/ermis/ /var/log/ermis/
ADD ermis-1.2.4-3.el7.cern.x86_64.rpm   .
RUN rpm -ivvh ermis-1.2.4-3.el7.cern.x86_64.rpm
EXPOSE 8080
ENTRYPOINT ["ermis"]
