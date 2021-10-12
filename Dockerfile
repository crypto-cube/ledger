FROM golang:1.17.2-alpine AS build

WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 go build -o /bin/numary

FROM scratch
COPY --from=build /bin/numary /bin/numary

EXPOSE 3068

ENTRYPOINT ["/bin/numary"]

