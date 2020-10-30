FROM cern/c8-base
WORKDIR /root/
LABEL maintainer="LB-Experts <lb-experts@cern.ch>"
RUN  mkdir -p  /var/lib/ermis/ /var/log/ermis/ && \
     dnf -y  install "dnf-command(config-manager)"
#     dnf config-manager --add-repo  http://linuxsoft.cern.ch/internal/repos/lb8-stable/x86_64/os  && \
#     ( yum install -y ermis || true)
EXPOSE 8080
CMD ["ermis", "-config=/usr/local/etc/goermis.yaml", "-home=/var/lib/ermis/"]


