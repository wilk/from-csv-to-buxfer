FROM python:3.6.2

RUN mkdir -p /opt/app
WORKDIR /opt/app

EXPOSE 8080

# custom entrypoint
COPY ./docker/entrypoint.sh /
RUN chmod 777 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
