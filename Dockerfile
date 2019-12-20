FROM golang:latest
WORKDIR /home/applywork
COPY wxreply /home/applywork
ENV PORT 8080
EXPOSE 8080

