FROM alpine
MAINTAINER Fabrice Aneche <akh@nobugware.com>

ADD ./regionagogo.linux /regionagogo

USER nobody
ENTRYPOINT /regionagogo

EXPOSE 8082
