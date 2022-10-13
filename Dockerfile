# syntax=docker/dockerfile:1
FROM golang:1.16 as base
WORKDIR /go/src/jirevwe/cascade

COPY ./go.mod /go/src/jirevwe/cascade
COPY ./go.sum /go/src/jirevwe/cascade

# Get dependancies - will also be cached if we don't change mod/sum
RUN go mod download
RUN go mod verify

# COPY the source code as the last step
COPY . .

RUN CGO_ENABLED=0
RUN go install ./cmd

FROM gcr.io/distroless/base
COPY --from=base /go/bin/cmd /

CMD [ "/cmd", "server", "--config", "./config.json" ]

EXPOSE 4400
