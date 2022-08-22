FROM golang:alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o main .


FROM alpine:latest

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/main .      

# Expose port 8080 to the outside world
EXPOSE 8089

#Command to run the executable
CMD ["./main"]

# docker build -t muthurajrprivaterepo/validate:v2
# docker push muthurajrprivaterepo/validate:v2