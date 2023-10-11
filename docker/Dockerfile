FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/hybrid.hybrid-scheduler \
    USER_UID=root \
    USER_NAME=discovery
COPY _output/bin/hybrid.hybrid-scheduler ${OPERATOR}


ENTRYPOINT ["/usr/local/bin/hybrid.hybrid-scheduler"]

USER ${USER_UID}