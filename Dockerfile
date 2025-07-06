# go_server/api/Dockerfile

# --- STAGE 1: Build da Aplicação Go ---
# MUDANÇA AQUI: Altere a versão do Go para 1.24-alpine (ou superior, se o go.mod exigir)
FROM golang:1.24-alpine AS builder 

WORKDIR /app

# Copia os arquivos go.mod e go.sum para baixar as dependências primeiro
COPY go.mod ./
COPY go.sum ./

# Baixa as dependências (fora do vendor)
RUN go mod download

# Copia o código-fonte da aplicação
COPY . .

# Compila a aplicação Go
# CGO_ENABLED=0 é importante para criar um binário estaticamente ligado (melhor para Alpine)
# -o server é o nome do executável de saída
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o server .

# --- STAGE 2: Imagem Final Leve ---
FROM alpine:latest

WORKDIR /root/

# Copia o binário compilado do estágio anterior
COPY --from=builder /app/server .

# Copia o arquivo .env (importante para carregar as variáveis de ambiente)
COPY ./.env ./.env 

# Permissões de execução para o binário
RUN chmod +x server

# Expõe a porta que sua aplicação Go vai escutar
EXPOSE 8080

# Comando para rodar a aplicação quando o container iniciar
CMD ["./server"]