FROM golang:1.14-alpine as build-stage
WORKDIR /app
COPY . .
RUN go build

FROM alpine

COPY --from=build-stage /app/tail-based-sampling ./
CMD [ "./tail-based-sampling" ] 
