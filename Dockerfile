FROM golang:1.16-alpine as build

ARG LOFTSMAN_VERSION=dev
ARG HELM_VERSION=3.5.4

ENV GO111MODULE=on

# We're actually going to build Helm from source in this container so that we have
# more assurance of the binary running stably within an alpine-based container image
RUN apk add --no-cache git make bash
RUN git clone https://github.com/helm/helm.git /helm
WORKDIR /helm
RUN git checkout v${HELM_VERSION}
RUN make

COPY . /loftsman
WORKDIR /loftsman

RUN go build -o ./loftsman -ldflags "-X 'github.com/Cray-HPE/loftsman/cmd.Version=${LOFTSMAN_VERSION}'"

FROM alpine:latest as release

COPY --from=build /helm/bin/helm /usr/bin/helm
COPY --from=build /loftsman/loftsman /usr/bin/loftsman
RUN chmod +x /usr/bin/helm
RUN chmod +x /usr/bin/loftsman
