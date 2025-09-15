# STAGE 1: BUILD
# Go compiler with Alpine base
FROM golang:1.23-alpine AS builder         
# Set build directory
WORKDIR /app                               
# Copy source code
COPY . .                                   
# Compile Go binary
RUN go build -o app main.go                

# STAGE 2: RUN
# Minimal runtime image
FROM alpine:latest                         
# Set runtime directory
WORKDIR /app                               
# Copy binary from builder
COPY --from=builder /app/app .             
# Copy web assets (templates/static)
COPY --from=builder /app/web ./web         
# Document port
EXPOSE 8080                                
# Start application
CMD ["./app"]