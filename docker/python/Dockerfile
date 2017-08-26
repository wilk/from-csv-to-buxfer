FROM python:3.6.2

RUN pip install "virtualenv==15.1.0"

RUN mkdir -p /opt/app
WORKDIR /opt/app

# custom entrypoint
COPY ./docker/python/entrypoint.sh /
RUN chmod 777 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
