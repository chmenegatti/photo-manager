# Photo Manager

Um gerenciador de fotos inspirado no Google Fotos, desenvolvido em Golang.

## Visão Geral

Este projeto visa criar uma aplicação web completa para gerenciar suas fotos, permitindo upload, organização automática por data, detecção de duplicatas, criação de álbuns e busca eficiente.

## Estrutura do Projeto

A arquitetura do projeto segue a seguinte estrutura de diretórios:

```bash
photo-manager/
├── cmd/                     # Ponto de entrada da aplicação
├── internal/                # Pacotes internos com a lógica de negócio
│   ├── api/                 # Handlers da API REST
│   ├── config/              # Configurações da aplicação
│   ├── database/            # Conexão e modelos do banco de dados
│   ├── exif/                # Funções para manipulação de EXIF
│   ├── storage/             # Funções para manipulação de arquivos
│   └── service/             # Lógica de negócio (camada de serviço)
├── pkg/                     # Pacotes utilitários e reutilizáveis
│   ├── utils/
│   └── logger/
├── web/                     # Arquivos estáticos da interface web
├── .env.example             # Exemplo de variáveis de ambiente
├── go.mod                   # Módulos Go
├── go.sum                   # Checksums dos módulos
└── README.md                # Documentação do projeto

```

## Como Executar (Localmente)

1. **Clone o repositório:**

    ```bash
    git clone [URL_DO_REPOSITORIO]
    cd photo-manager
    ```

    (Obs: Substitua `[URL_DO_REPOSITORIO]` pela URL do seu repositório Git, caso esteja usando um)

2. **Instale as dependências:**

    ```bash
    go mod tidy
    ```

3. **Execute a aplicação:**

    ```bash
    go run cmd/main.go
    ```

    A aplicação estará disponível em `http://localhost:8080`. Você pode testar a rota de exemplo acessando `http://localhost:8080/ping`.

## Funcionalidades Planejadas

* Upload de fotos via API REST.
* Extração automática de metadados EXIF e organização por ano/mês.
* Detecção de fotos duplicadas.
* Sistema de álbuns personalizados.
* Busca eficiente por fotos.
* Interface web para visualização.
* API REST documentada.

---

#### `.env.example`

```dotenv
APP_PORT=8080
DATABASE_URL=./data/photo_manager.db # Exemplo para SQLite
PHOTO_STORAGE_PATH=./data/photos
```

---

### Verificando a Instalação

Após seguir esses passos, você pode verificar a estrutura de diretórios com o comando `ls -R`.

Para garantir que tudo está funcionando, vamos instalar a dependência do Gin e rodar o projeto:

1. **Instale o Gin Gonic:**

    ```bash
    go get github.com/gin-gonic/gin
    ```

2. **Execute a aplicação:**

    ```bash
    go run cmd/main.go
    ```

Você deverá ver no terminal a mensagem `Servidor iniciado na porta 8080`. Abrindo seu navegador e acessando `http://localhost:8080/ping`, você deverá ver o JSON `{"message":"pong"}`.
