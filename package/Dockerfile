FROM alpine:3.12.3
ARG ARCH
ENV ARCH=$ARCH
COPY bin/drone-docker-image-digests-linux-${ARCH} /usr/bin/drone-docker-image-digests
CMD ["drone-docker-image-digests"]
