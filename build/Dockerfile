FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/hcp-scheduler \
    USER_UID=1001 \
    USER_NAME=hcp-scheduler

# install operator binary
COPY _output/bin/hcp-scheduler ${OPERATOR}

COPY bin /usr/local/bin

RUN  /usr/local/bin/user_setup

RUN chmod +x /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
