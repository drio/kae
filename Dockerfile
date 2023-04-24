FROM alpine:latest
WORKDIR /app
COPY . ./
CMD [ "./main-linux-amd64", "--delaySecs", ${KAE_DELAY_SECS}" ]
