FROM cern/cc7-base

RUN yum install -y \
    httpd mod_ssl \
    mod_auth_kerb \
    mod_wsgi \
    opensaml-schemas && \
    mkdir -p /var/www/ermis && \
    yum clean all 

#COPY httpd/httpd.conf /etc/httpd/conf/
COPY httpd/11-goermis.cern.ch.conf /etc/httpd/conf.d/
COPY httpd/25-https-to-https-redirection.conf /etc/httpd/conf.d/
COPY /entr.sh /
RUN chmod +x entr.sh
EXPOSE 80
EXPOSE 443
CMD ["/entr.sh"]




