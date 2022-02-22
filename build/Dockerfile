FROM gcr.io/distroless/static:nonroot

ENV OPERATOR=/usr/local/bin/rabbitmq-operator

# install operator binary
COPY build/_output/bin/rabbitmq-operator ${OPERATOR}

ENTRYPOINT ["${OPERATOR}"]
