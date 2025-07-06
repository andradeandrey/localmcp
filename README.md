# Servidor MCP GitHub

Este é um servidor MCP (Model Context Protocol) que se integra com a API do GitHub, permitindo que assistentes de IA interajam com repositórios, issues, pull requests e outros recursos do GitHub.

## Configuração

### 1. Instalar Go

Certifique-se de ter Go instalado (versão 1.19 ou superior):

```bash
go version
```

### 2. Token do GitHub

1. Acesse [GitHub Personal Access Tokens](https://github.com/settings/tokens)
2. Clique em "Generate new token (classic)"
3. Selecione as seguintes permissões:
   - `repo` (acesso total aos repositórios)
   - `user` (acesso ao perfil do usuário)
   - `read:org` (se precisar acessar organizações)
4. Copie o token gerado

### 3. Configurar variável de ambiente

```bash
export GITHUB_TOKEN="seu_token_aqui"
```

### 4. Compilar e executar

```bash
go build -o mcp-github-server main.go
./mcp-github-server
```

## Funcionalidades

O servidor MCP fornece as seguintes ferramentas:

### 1. `get_user`
Obter informações de um usuário GitHub.

**Parâmetros:**
- `username` (opcional): Nome do usuário. Se não fornecido, retorna info do usuário autenticado.

**Exemplo de uso:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_user",
    "arguments": {
      "username": "octocat"
    }
  }
}
```

### 2. `get_repos`
Listar repositórios de um usuário.

**Parâmetros:**
- `username` (opcional): Nome do usuário. Se não fornecido, lista repos do usuário autenticado.

### 3. `get_issues`
Listar issues de um repositório.

**Parâmetros:**
- `owner` (obrigatório): Proprietário do repositório
- `repo` (obrigatório): Nome do repositório

### 4. `get_pull_requests`
Listar pull requests de um repositório.

**Parâmetros:**
- `owner` (obrigatório): Proprietário do repositório
- `repo` (obrigatório): Nome do repositório

### 5. `get_commits`
Listar commits de um repositório.

**Parâmetros:**
- `owner` (obrigatório): Proprietário do repositório
- `repo` (obrigatório): Nome do repositório

### 6. `get_content`
Obter conteúdo de um arquivo no repositório.

**Parâmetros:**
- `owner` (obrigatório): Proprietário do repositório
- `repo` (obrigatório): Nome do repositório
- `path` (obrigatório): Caminho do arquivo

## Protocolo MCP

O servidor implementa o protocolo MCP versão 2024-11-05. Ele se comunica através de stdin/stdout usando mensagens JSON-RPC.

### Mensagens suportadas:

1. **initialize**: Inicializa o servidor
2. **tools/list**: Lista todas as ferramentas disponíveis
3. **tools/call**: Executa uma ferramenta específica
4. **ping**: Verifica se o servidor está ativo

### Exemplo de inicialização:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    }
  }
}
```

## Integração com Claude Desktop

Para usar este servidor com Claude Desktop, adicione a seguinte configuração ao seu arquivo de configuração:

```json
{
  "mcpServers": {
    "github": {
      "command": "./mcp-github-server",
      "env": {
        "GITHUB_TOKEN": "seu_token_aqui"
      }
    }
  }
}
```

## Arquivo Makefile

Crie um `Makefile` para facilitar o build e execução:

```makefile
# Makefile para o servidor MCP GitHub

.PHONY: build run clean test

# Variáveis
BINARY_NAME=mcp-github-server
MAIN_FILE=main.go

# Build do projeto
build:
  go build -o $(BINARY_NAME) $(MAIN_FILE)

# Executar o servidor
run: build
  ./$(BINARY_NAME)

# Executar com verificação de token
run-check: build
  @if [ -z "$(GITHUB_TOKEN)" ]; then \
    echo "Erro: GITHUB_TOKEN não definido"; \
    echo "Execute: export GITHUB_TOKEN='seu_token_aqui'"; \
    exit 1; \
  fi
  ./$(BINARY_NAME)

# Limpar arquivos compilados
clean:
  rm -f $(BINARY_NAME)

# Testar o servidor
test:
  go test -v ./...

# Instalar dependências (se houver)
deps:
  go mod tidy

# Verificar formatação
fmt:
  go fmt ./...

# Verificar código
vet:
  go vet ./...

# Build para diferentes plataformas
build-all:
  GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
  GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
  GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
  GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
```

## Script de Teste

Crie um script para testar o servidor:

```bash
#!/bin/bash
# test-server.sh

# Verificar se o token está definido
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Erro: GITHUB_TOKEN não definido"
    echo "Execute: export GITHUB_TOKEN='seu_token_aqui'"
    exit 1
fi

# Compilar o servidor
echo "Compilando servidor..."
go build -o mcp-github-server main.go

# Iniciar o servidor em background
echo "Iniciando servidor..."
./mcp-github-server &
SERVER_PID=$!

# Aguardar um pouco para o servidor iniciar
sleep 2

# Testar inicialização
echo "Testando inicialização..."
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./mcp-github-server

# Testar ping
echo "Testando ping..."
echo '{"jsonrpc":"2.0","id":2,"method":"ping","params":{}}' | ./mcp-github-server

# Testar lista de ferramentas
echo "Testando lista de ferramentas..."
echo '{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}' | ./mcp-github-server

# Finalizar servidor
kill $SERVER_PID
echo "Teste concluído!"
```

## Exemplo de Uso Completo

Aqui está um exemplo completo de como usar o servidor:

```bash
# 1. Definir token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# 2. Compilar
go build -o mcp-github-server main.go

# 3. Executar
./mcp-github-server
```

Então, em outro terminal, você pode enviar comandos:

```bash
# Inicializar
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./mcp-github-server

# Listar ferramentas
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./mcp-github-server

# Obter informações do usuário
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_user","arguments":{"username":"octocat"}}}' | ./mcp-github-server

# Listar repositórios
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_repos","arguments":{"username":"octocat"}}}' | ./mcp-github-server

# Listar issues de um repositório
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_issues","arguments":{"owner":"facebook","repo":"react"}}}' | ./mcp-github-server
```

## Recursos Avançados

### Autenticação
O servidor usa tokens de acesso pessoal do GitHub para autenticação. Certifique-se de que o token tenha as permissões necessárias para acessar os recursos desejados.

### Rate Limiting
O servidor respeita os limites de taxa da API do GitHub. Se você encontrar erros de limite de taxa, aguarde antes de fazer mais solicitações.

### Tratamento de Erros
O servidor inclui tratamento robusto de erros, retornando códigos de erro JSON-RPC apropriados quando algo dá errado.

### Extensibilidade
O código foi estruturado para facilitar a adição de novas funcionalidades. Você pode adicionar novos métodos à API do GitHub modificando:

1. Adicionar nova estrutura de dados (se necessário)
2. Implementar método no `GitHubClient`
3. Adicionar nova ferramenta ao array `tools`
4. Implementar handler no `MCPServer`

## Estrutura do Projeto

```
mcp-github-server/
├── main.go              # Código principal
├── Makefile            # Automação de build
├── test-server.sh      # Script de teste
├── README.md           # Documentação
└── go.mod              # Módulo Go (se necessário)
```

## Solução de Problemas

### Token inválido
- Verifique se o token está correto
- Confirme que o token tem as permissões necessárias
- Teste o token manualmente com curl

### Servidor não responde
- Verifique se o servidor está rodando
- Confirme que as mensagens JSON estão bem formadas
- Verifique os logs de erro

### Erro de API
- Verifique se o repositório/usuário existe
- Confirme que você tem acesso ao recurso solicitado
- Verifique limites de taxa da API

## Licença

Este projeto é fornecido como exemplo educacional. Use sob sua própria responsabilidade.