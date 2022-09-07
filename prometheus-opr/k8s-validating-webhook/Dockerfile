FROM golang:alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o /main .


FROM alpine:latest

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /main /main      

# Expose port 8080 to the outside world
EXPOSE 8089
EXPOSE 443

VOLUME /etc/webhook
#Command to run the executable
#CMD ["./main"]
ENTRYPOINT ["/main"]

# docker build -t muthurajrprivaterepo/validate:v0.8 .
# docker push muthurajrprivaterepo/validate:v0.8