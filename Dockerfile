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

COPY modules.sh /opt/modules.sh

RUN set -xe && \
	/opt/modules.sh

