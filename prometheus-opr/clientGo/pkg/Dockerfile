FROM debian
COPY ./app /app
ENTRYPOINT /app

# export GO111MODULE=on;
# go mod tidy
# GOOS=linux go build -o ./app .
# docker build -t muthurajrprivaterepo/sampleclientgo:v5.5 .
# docker push muthurajrprivaterepo/sampleclientgo:v5.5
# minikube image build -t muthurajrprivaterepo/sampleclientgo:v3.5 .

# FROM golang AS builder

# WORKDIR $GOPATH/src/clientgoapp
# COPY . ./
# RUN GOPROXY=direct CGO_ENABLED=0 GOOS=linux go build -o /app .

# FROM scratch
# COPY --from=builder /app /app
# EXPOSE 8888
# ENTRYPOINT /app