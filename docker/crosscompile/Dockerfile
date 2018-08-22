FROM dockercore/golang-cross:1.10.3
MAINTAINER Yusaku Hatanaka (hatappi)

# install amd64 packages
RUN apt-get update -y && \
    apt-get install -y libasound2-dev

# install 386 packages
RUN dpkg --add-architecture i386 && \
    apt-get update -y && \
    apt-get install -y g++-multilib libasound2-dev:i386
