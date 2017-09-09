FROM golang:1.9.0

RUN curl https://glide.sh/get | sh

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# custom entrypoint
COPY ./docker/entrypoint.sh /
RUN chmod 777 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
