FROM alpine 
MAINTAINER Fabrice Aneche <akh@nobugware.com>

RUN mkdir /app
ADD regionagogo.linux region.db /app/
EXPOSE 8082
EXPOSE 8083

CMD ["-dbpath", "/app/region.db"]
ENTRYPOINT ["/app/regionagogo.linux"]

