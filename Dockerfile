FROM golang:alpine AS Builder
WORKDIR /home/icapeg
COPY . .
RUN go build .

FROM alpine:3.18.3
WORKDIR /home/icapeg
# RUN apk add --no-cache libc6-compat
RUN apk --no-cache add ca-certificates libc6-compat
COPY --from=Builder ./home/icapeg/icapeg .
COPY --from=Builder ./home/icapeg/config.toml .
COPY --from=Builder ./home/icapeg/block-page.html .

EXPOSE 1344
ENTRYPOINT ["./icapeg"]
