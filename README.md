# 📦 Alfredo Storage Manager (API Backend)

> Um servidor de armazenamento de arquivos (File Manager API) extremamente leve, otimizado para rodar em hardwares legados (32-bits) e focado no consumo mínimo de memória RAM.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Architecture](https://img.shields.io/badge/Arch-32--bit-red?style=for-the-badge)
![Nginx](https://img.shields.io/badge/Nginx-Reverse%20Proxy-009639?style=for-the-badge&logo=nginx)

## 🚀 Sobre o Projeto
O backend do **Alfredo Storage Manager** foi projetado com um objetivo radical: gerenciar um sistema de arquivos completo utilizando o menor *footprint* de memória possível, sem sacrificar estabilidade. Para isso, o sistema trabalha de forma puramente *headless* (apenas API JSON), delegando a entrega de assets estáticos (Frontend) para o Nginx.

## ⚡ Otimizações de Performance (Low RAM Constraints)
- **Zero-Allocation Programming**: Uso massivo de `sync.Pool` para reutilizar buffers na memória em rotas de I/O de alta densidade, reduzindo a sobrecarga do *Garbage Collector*.
- **Stream I/O & Buffered Writers**: Evitamos carregar arquivos inteiros para a RAM. Utilizamos buffers estáticos (ex: 32KB) com `bufio` e rotas baseadas em streams sob demanda (`io.CopyBuffer` e `MultipartReader`).
- **Binary Stripping**: Compilado com `-ldflags="-s -w"` e arquitetura apontada para CPUs legadas (`GOARCH=386`).
- **Monitoramento Estrito**: Timeouts restritivos (`Read`, `Write`, `Idle`) mitigam ataques parasitas e vazamento de *goroutines*.
- **Pprof Integrado**: Rota embutida para monitoramento ativo e mapeamento de memória em tempo real da aplicação (`/debug/pprof`).

## ⚙️ Instalação e Build

### 1. Requisitos
- Go 1.21+
- Nginx (recomendado para atuar como proxy reverso)

### 2. Compilação Otimizada para 32-bits
```bash
GOARCH=386 go build -ldflags="-s -w" -o server32_optimized main.go
```

## 🔗 Endpoints da API
Rotas protegidas que manipulam o sistema de arquivos base:
- `GET /files` - Lista arquivos e pastas
- `POST /create-folder` - Cria novos diretórios
- `POST /upload` - Processa streaming de arquivos 
- `GET /download` - Faz o *sendfile* do conteúdo com consumo mínimo de memória
- `DELETE /delete` - Remove itens
- `POST /rename` - Renomeia arquivos e pastas

*Nota: O frontend desta aplicação vive em seu próprio repositório: [AlfredoStorageManagerFront](https://github.com/Aaron-GMM/AlfredoStorageManagerFront).*
