FROM golang:1.9.0

RUN curl https://glide.sh/get | sh

RUN mkdir -p /opt/app
WORKDIR /opt/app

# custom entrypoint
COPY ./docker/golang/entrypoint.sh /
RUN chmod 777 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
