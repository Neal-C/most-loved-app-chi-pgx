ARG APP_NAME=app

# Build stage
FROM golang:1.21.3 as build
ARG APP_NAME
ENV APP_NAME=$APP_NAME
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /$APP_NAME

# Production stage
FROM alpine:latest as production
ARG APP_NAME
ENV APP_NAME=$APP_NAME
WORKDIR /root/
COPY --from=build /$APP_NAME ./
CMD ./$APP_NAME



## ! WORKS!
# # Build stage
# FROM golang:1.21.3 as build
# WORKDIR /app
# COPY . .
# RUN go mod download
# RUN CGO_ENABLED=0 go build -o main main.go

# # Production stage
# FROM alpine:latest as production
# WORKDIR /app
# COPY --from=build /app/main ./

# EXPOSE 9000
# CMD [ "/app/main" ]
