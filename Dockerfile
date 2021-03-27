# build stage
FROM golang:1.16.2-alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /src
RUN cd /src && go build -o app

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/app /app/
ENTRYPOINT ./app