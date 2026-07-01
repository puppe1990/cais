# Stage 1: Tailwind CSS
FROM node:22-alpine AS css
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY input.css tailwind.config.js ./
COPY web/templates ./web/templates
RUN npx tailwindcss -i input.css -o styles.css --minify

# Stage 2: Go build
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=css /app/styles.css ./web/static/css/styles.css
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /cais ./cmd/server

# Stage 3: Minimal runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1000 cais
WORKDIR /app
COPY --from=build /cais /app/cais
COPY --from=build /app/web/static /app/web/static
RUN mkdir -p /app/data && chown -R cais:cais /app
USER cais
EXPOSE 8080
ENV PORT=:8080
ENV DB_PATH=/app/data/app.db
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/cais"]