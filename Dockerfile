# syntax=docker/dockerfile:1.4

###########################
# 🔨 STAGE: Builder
###########################

# L’immagine base viene ancora passata da build.sh
ARG BASE_IMAGE
FROM ${BASE_IMAGE:-golang:1.21-bullseye} AS builder

# Rimuoviamo i vecchi ARG GOOS/GOARCH/GOARM
# Aggiungiamo invece TARGETPLATFORM
ARG TARGETPLATFORM
ENV CGO_ENABLED=0

# Impostiamo la cartella di lavoro e copiamo sorgenti
WORKDIR /app
COPY . .

# Selezioniamo i valori di GOOS/GOARCH/GOARM in base a TARGETPLATFORM
RUN echo "Building for $TARGETPLATFORM" \
 && case "$TARGETPLATFORM" in \
      "linux/arm/v6") GOOS=linux GOARCH=arm GOARM=6 ;; \
      "linux/arm/v7") GOOS=linux GOARCH=arm GOARM=7 ;; \
      "linux/amd64") GOOS=linux GOARCH=amd64 ;; \
      *)             GOOS=linux GOARCH=$TARGETARCH ;; \
    esac \
 && go build -trimpath -ldflags "-s -w" -o meshspy ./cmd/meshspy

###########################
# 🏁 STAGE: Runtime finale
###########################

FROM alpine:3.18 AS runtime

WORKDIR /app
COPY --from=builder /app/meshspy .

ENTRYPOINT ["./meshspy"]
