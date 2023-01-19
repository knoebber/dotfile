# fly.io uses Ubuntu/Debian instead of Alpine to avoid DNS resolution issues in production.

ARG BUILDER_IMAGE="golang:1.18.10-buster"
ARG RUNNER_IMAGE="debian:bullseye-20220801-slim"

FROM ${BUILDER_IMAGE} as builder

# prepare build dir
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN make dotfilehub

# start a new build stage so that the final image will only contain the compiled binary
FROM ${RUNNER_IMAGE}

RUN apt-get update -y && apt-get install -y libstdc++6 openssl libncurses5 locales \
  && apt-get clean && rm -f /var/lib/apt/lists/*_*

# Set the locale
RUN sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen && locale-gen

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

WORKDIR "/app/bin"
RUN chown nobody /app

COPY --from=builder --chown=nobody:root /app/bin/dotfilehub .

USER nobody

CMD [\
    "/app/bin/dotfilehub",\
    "-addr=:8080",\
    "-db=/data/dotfilehub.db",\
    "-host=dotfilehub.com",\
    "-secure",\
    "-proxyheaders"\
]
