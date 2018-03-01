FROM golang:1.9-alpine 
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --virtual git
RUN apk add --no-cache --virtual make 
RUN apk add --no-cache --virtual bash 

CMD ["bash"]
