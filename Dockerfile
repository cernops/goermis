# THIS IS A TEMPORARY SOLUTION FOR KEEPING THE DOCKER IMAGES UP-TO-DATE
# THE PERMANENT SOLUTION IS AT THE END AND CAN BE USED ONCE THE RPM FILE IS ON THE REPO

FROM cern/c8-base
LABEL maintainer="LB-Experts <lb-experts@cern.ch>"
RUN   yum -y install golang git 
COPY .  .
RUN go mod download
RUN go build -o goermis .
RUN  mkdir -p  /var/lib/ermis/ /var/log/ermis/
COPY staticfiles      /var/lib/ermis/staticfiles
COPY templates        /var/lib/ermis/templates
EXPOSE 8080
ENTRYPOINT ["/goermis"]
CMD ["-config=/usr/local/etc/goermis.yaml", "-home=/var/lib/ermis/"]





# THIS IS THE SOLUTION TO BE USED WITH THE RPM. WHEN THE RPM IS IN THE REPO UNCOMMENT THE LINES WITH 2 # , TOO.
# OTHERWISE, FOR A DUMMY DOCKER_BUILD, UNCOMMENT ONLY THE LINES WITH 1 #

#FROM cern/c8-base
#WORKDIR /root/
#LABEL maintainer="LB-Experts <lb-experts@cern.ch>"
#RUN  mkdir -p  /var/lib/ermis/ /var/log/ermis/ && \
#     dnf -y  install "dnf-command(config-manager)"
##     dnf config-manager --add-repo  http://linuxsoft.cern.ch/internal/repos/lb8-stable/x86_64/os  && \
##    ( yum install -y ermis || true)
#EXPOSE 8080
#CMD ["ermis", "-config=/usr/local/etc/goermis.yaml", "-home=/var/lib/ermis/"]
