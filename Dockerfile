FROM golang:1.10-stretch

RUN set -xe && \
	rm /etc/localtime && \
	ln -s /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
	echo 'Asia/Tokyo' > /etc/timezone


RUN set -xe && \
	export DEBIAN_FRONTEND=noninteractive && \
	apt-get update && \
	apt-get install -y --no-install-recommends \
		upx-ucl && \
	rm -rf /var/lib/apt/lists/*

COPY bin/get-external-modules.sh /opt/get-external-modules.sh

RUN set -xe && \
	/opt/get-external-modules.sh

