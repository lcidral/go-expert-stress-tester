FROM golang:1.23-alpine3.19 AS builder

# Configurar os flags de compilação para otimização
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Configurar o diretório de trabalho
WORKDIR /build

# Copiar os arquivos de dependências primeiro
COPY go.* ./
RUN go mod download

# Copiar o código fonte
COPY . .

# Compilar a aplicação com otimizações
RUN go build \
    -ldflags="-w -s" \
    -trimpath \
    -o stress-tester \
    ./cmd/main.go

# Estágio final
FROM alpine:3.19 AS final

# Adicionar certificados CA e timezone data
RUN apk --no-cache add \
    ca-certificates \
    tzdata

# Criar usuário não-root
RUN adduser -D appuser

# Configurar o diretório da aplicação
WORKDIR /app

# Copiar o binário compilado
COPY --from=builder /build/stress-tester .

# Definir o proprietário dos arquivos
RUN chown -R appuser:appuser /app

# Mudar para usuário não-root
USER appuser

# Configurar o entrypoint
ENTRYPOINT ["/app/stress-tester"]