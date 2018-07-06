FROM circleci/buildpack-deps:bionic

RUN sudo apt-get update && \
  sudo apt-get install -y --no-install-recommends curl gnupg && \
  echo "deb http://packages.cloud.google.com/apt cloud-sdk-bionic main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
  curl http://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add - && \
  sudo apt-get update && \
  sudo apt-get install -y --no-install-recommends \
    python \
    golang-1.9-go \
    google-cloud-sdk \
    google-cloud-sdk-app-engine-go
ENV PATH $PATH:/usr/lib/go-1.9/bin
