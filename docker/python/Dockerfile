FROM python:3.6.2

RUN mkdir -p /opt/app
RUN mkdir -p /opt/samples
WORKDIR /opt

# custom entrypoint
COPY ./docker/python/entrypoint.sh /
RUN chmod 777 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
