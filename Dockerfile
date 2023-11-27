FROM  golang:1.21.4-alpine as build
WORKDIR /go/src
RUN apk add --no-cache make git && \
    git clone https://github.com/nuclear06/spot && \
    cd spot && \
    make

FROM gcr.io/distroless/static-debian11:latest
EXPOSE 2023
WORKDIR /spot
COPY --from=build /go/src/spot/spot /spot/app
RUN ["/spot/app" ,"conf", "-i"]
ENTRYPOINT  ["/spot/app"]