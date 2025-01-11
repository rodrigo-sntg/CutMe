# syntax=docker/dockerfile:1

##
# Fase 1: Build da aplicação
##
FROM golang:1.23-alpine AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de dependências para fazer cache do go.mod/go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o restante do código-fonte
COPY . .

# Compila o binário
RUN go build -o /go/bin/app cmd/main.go
RUN chmod +x /go/bin/app

##
# Fase 2: Imagem de runtime
##
FROM alpine:3.17

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia o binário gerado na fase anterior
COPY --from=builder /go/bin/app /app/app

# Expõe a porta da aplicação
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["/app/app"]
