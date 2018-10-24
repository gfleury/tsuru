FROM golang:1.11-alpine 
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --virtual git
RUN apk add --no-cache --virtual make 
RUN apk add --no-cache --virtual bash 
RUN apk add gcc musl-dev --no-cache

CMD ["bash"]
