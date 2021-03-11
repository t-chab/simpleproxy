# Build simple-proxy
FROM golang:1.16 AS builder

ARG GIT_REPO="https://github.com/tchabaud/simple-proxy.git"
RUN git clone ${GIT_REPO} /app
WORKDIR /app
RUN GOCACHE=/tmp make

