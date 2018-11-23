FROM circleci/buildpack-deps:bionic

RUN sudo apt-get update && \
  sudo apt-get install -y software-properties-common && \
  sudo add-apt-repository ppa:longsleep/golang-backports && \
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add - && \
  sudo add-apt-repository "deb http://packages.cloud.google.com/apt cloud-sdk-bionic main" && \
  sudo apt-get install -y --no-install-recommends \
    golang-1.11 \
    python \
    google-cloud-sdk \
    google-cloud-sdk-app-engine-python \
    google-cloud-sdk-app-engine-go \
    google-cloud-sdk-datastore-emulator
ENV PATH $PATH:/usr/lib/go-1.11/bin
