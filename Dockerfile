FROM golang:1.12.3 AS BUILD

RUN mkdir /wfs-eye
WORKDIR /wfs-eye

ADD go.mod .
ADD go.sum .
RUN go mod download

#now build source code
ADD . ./
RUN go build -o /go/bin/wfs-eye



FROM golang:1.12.3

ENV WFS3_API_URL ''
ENV LOG_LEVEL 'info'
ENV MONGO_DBNAME=admin
ENV MONGO_ADDRESS=mongo
ENV MONGO_USERNAME=root
ENV MONGO_PASSWORD=root

COPY --from=BUILD /go/bin/* /bin/
ADD /startup.sh /
ENTRYPOINT /startup.sh

EXPOSE 4000

