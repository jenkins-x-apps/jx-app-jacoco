FROM maven:3.6.1-ibmjava-alpine

RUN apk --update add ca-certificates

COPY ./build/jx-app-jacoco /jx-app-jacoco

EXPOSE 8080
ENTRYPOINT ["/jx-app-jacoco"]

