FROM ubuntu:18.04

RUN apt-get update && \
  apt-get install -y --no-install-recommends curl gnupg && \
  echo "deb http://packages.cloud.google.com/apt cloud-sdk-bionic main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
  curl http://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
  apt-get update && \
  apt-get install -y --no-install-recommends \
    golang-1.9 \
    google-cloud-sdk \
    google-cloud-sdk-app-engine-go
