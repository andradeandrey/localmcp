package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Estruturas MCP
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]interface{} `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]interface{} `json:"serverInfo"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type CallToolResult struct {
	Content []map[string]interface{} `json:"content"`
}

// Estruturas GitHub API
type GitHubRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type GitHubUser struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Company   string `json:"company"`
	Email     string `json:"email"`
	HTMLURL   string `json:"html_url"`
	Followers int    `json:"followers"`
	Following int    `json:"following"`
}

type GitHubIssue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GitHubPR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GitHubCommit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"author"`
	HTMLURL string `json:"html_url"`
}

type GitHubContent struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int    `json:"size"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	HTMLURL  string `json:"html_url"`
}

// Cliente GitHub
type GitHubClient struct {
	token   string
	baseURL string
	client  *http.Client
}

func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token:   token,
		baseURL: "https://api.github.com",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (gc *GitHubClient) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, gc.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+gc.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "MCP-GitHub-Server/1.0")

	return gc.client.Do(req)
}

func (gc *GitHubClient) GetUser(ctx context.Context, username string) (*GitHubUser, error) {
	endpoint := "/users/" + username
	if username == "" {
		endpoint = "/user"
	}

	resp, err := gc.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (gc *GitHubClient) GetRepos(ctx context.Context, username string) ([]GitHubRepo, error) {
	endpoint := "/users/" + username + "/repos"
	if username == "" {
		endpoint = "/user/repos"
	}

	resp, err := gc.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (gc *GitHubClient) GetIssues(ctx context.Context, owner, repo string) ([]GitHubIssue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)

	resp, err := gc.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
}

func (gc *GitHubClient) GetPullRequests(ctx context.Context, owner, repo string) ([]GitHubPR, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)

	resp, err := gc.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var prs []GitHubPR
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	return prs, nil
}

func (gc *GitHubClient) GetCommits(ctx context.Context, owner, repo string) ([]GitHubCommit, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/commits", owner, repo)

	resp, err := gc.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var commits []GitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	return commits, nil
}

func (gc *GitHubClient) GetContent(ctx context.Context, owner, repo, path string) (*GitHubContent, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)

	resp, err := gc.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var content GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, err
	}

	return &content, nil
}

// Servidor MCP
type MCPServer struct {
	github *GitHubClient
	tools  []Tool
}

func NewMCPServer(token string) *MCPServer {
	return &MCPServer{
		github: NewGitHubClient(token),
		tools: []Tool{
			{
				Name:        "get_user",
				Description: "Obter informações de um usuário GitHub",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]interface{}{
							"type":        "string",
							"description": "Nome do usuário (deixe vazio para usuário autenticado)",
						},
					},
				},
			},
			{
				Name:        "get_repos",
				Description: "Listar repositórios de um usuário",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]interface{}{
							"type":        "string",
							"description": "Nome do usuário (deixe vazio para usuário autenticado)",
						},
					},
				},
			},
			{
				Name:        "get_issues",
				Description: "Listar issues de um repositório",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"owner": map[string]interface{}{
							"type":        "string",
							"description": "Proprietário do repositório",
						},
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Nome do repositório",
						},
					},
					"required": []string{"owner", "repo"},
				},
			},
			{
				Name:        "get_pull_requests",
				Description: "Listar pull requests de um repositório",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"owner": map[string]interface{}{
							"type":        "string",
							"description": "Proprietário do repositório",
						},
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Nome do repositório",
						},
					},
					"required": []string{"owner", "repo"},
				},
			},
			{
				Name:        "get_commits",
				Description: "Listar commits de um repositório",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"owner": map[string]interface{}{
							"type":        "string",
							"description": "Proprietário do repositório",
						},
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Nome do repositório",
						},
					},
					"required": []string{"owner", "repo"},
				},
			},
			{
				Name:        "get_content",
				Description: "Obter conteúdo de um arquivo no repositório",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"owner": map[string]interface{}{
							"type":        "string",
							"description": "Proprietário do repositório",
						},
						"repo": map[string]interface{}{
							"type":        "string",
							"description": "Nome do repositório",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Caminho do arquivo",
						},
					},
					"required": []string{"owner", "repo", "path"},
				},
			},
		},
	}
}

func (s *MCPServer) HandleMessage(ctx context.Context, msg MCPMessage) MCPMessage {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolsCall(ctx, msg)
	case "ping":
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  map[string]interface{}{"status": "pong"},
		}
	default:
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *MCPServer) handleInitialize(msg MCPMessage) MCPMessage {
	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			ServerInfo: map[string]interface{}{
				"name":    "GitHub MCP Server",
				"version": "1.0.0",
			},
		},
	}
}

func (s *MCPServer) handleToolsList(msg MCPMessage) MCPMessage {
	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"tools": s.tools,
		},
	}
}

func (s *MCPServer) handleToolsCall(ctx context.Context, msg MCPMessage) MCPMessage {
	var params CallToolParams
	
	// Converter params para JSON e depois fazer unmarshal
	paramsBytes, err := json.Marshal(msg.Params)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}
	
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	switch params.Name {
	case "get_user":
		return s.handleGetUser(ctx, msg, params)
	case "get_repos":
		return s.handleGetRepos(ctx, msg, params)
	case "get_issues":
		return s.handleGetIssues(ctx, msg, params)
	case "get_pull_requests":
		return s.handleGetPullRequests(ctx, msg, params)
	case "get_commits":
		return s.handleGetCommits(ctx, msg, params)
	case "get_content":
		return s.handleGetContent(ctx, msg, params)
	default:
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}
}

func (s *MCPServer) handleGetUser(ctx context.Context, msg MCPMessage, params CallToolParams) MCPMessage {
	username, _ := params.Arguments["username"].(string)

	user, err := s.github.GetUser(ctx, username)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: CallToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Usuário: %s\nNome: %s\nBio: %s\nLocalização: %s\nEmpresa: %s\nSeguidores: %d\nSeguindo: %d\nURL: %s",
						user.Login, user.Name, user.Bio, user.Location, user.Company, user.Followers, user.Following, user.HTMLURL),
				},
			},
		},
	}
}

func (s *MCPServer) handleGetRepos(ctx context.Context, msg MCPMessage, params CallToolParams) MCPMessage {
	username, _ := params.Arguments["username"].(string)

	repos, err := s.github.GetRepos(ctx, username)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Repositórios (%d):\n\n", len(repos)))
	for _, repo := range repos {
		result.WriteString(fmt.Sprintf("- %s\n", repo.Name))
		result.WriteString(fmt.Sprintf("  Descrição: %s\n", repo.Description))
		result.WriteString(fmt.Sprintf("  Privado: %t\n", repo.Private))
		result.WriteString(fmt.Sprintf("  URL: %s\n", repo.HTMLURL))
		result.WriteString("\n")
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: CallToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": result.String(),
				},
			},
		},
	}
}

func (s *MCPServer) handleGetIssues(ctx context.Context, msg MCPMessage, params CallToolParams) MCPMessage {
	owner, _ := params.Arguments["owner"].(string)
	repo, _ := params.Arguments["repo"].(string)

	issues, err := s.github.GetIssues(ctx, owner, repo)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Issues do %s/%s (%d):\n\n", owner, repo, len(issues)))
	for _, issue := range issues {
		result.WriteString(fmt.Sprintf("- #%d: %s\n", issue.Number, issue.Title))
		result.WriteString(fmt.Sprintf("  Estado: %s\n", issue.State))
		result.WriteString(fmt.Sprintf("  URL: %s\n", issue.HTMLURL))
		result.WriteString("\n")
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: CallToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": result.String(),
				},
			},
		},
	}
}

func (s *MCPServer) handleGetPullRequests(ctx context.Context, msg MCPMessage, params CallToolParams) MCPMessage {
	owner, _ := params.Arguments["owner"].(string)
	repo, _ := params.Arguments["repo"].(string)

	prs, err := s.github.GetPullRequests(ctx, owner, repo)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pull Requests do %s/%s (%d):\n\n", owner, repo, len(prs)))
	for _, pr := range prs {
		result.WriteString(fmt.Sprintf("- #%d: %s\n", pr.Number, pr.Title))
		result.WriteString(fmt.Sprintf("  Estado: %s\n", pr.State))
		result.WriteString(fmt.Sprintf("  URL: %s\n", pr.HTMLURL))
		result.WriteString("\n")
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: CallToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": result.String(),
				},
			},
		},
	}
}

func (s *MCPServer) handleGetCommits(ctx context.Context, msg MCPMessage, params CallToolParams) MCPMessage {
	owner, _ := params.Arguments["owner"].(string)
	repo, _ := params.Arguments["repo"].(string)

	commits, err := s.github.GetCommits(ctx, owner, repo)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commits do %s/%s (%d):\n\n", owner, repo, len(commits)))
	for _, commit := range commits {
		result.WriteString(fmt.Sprintf("- %s\n", commit.SHA[:7]))
		result.WriteString(fmt.Sprintf("  Mensagem: %s\n", commit.Message))
		result.WriteString(fmt.Sprintf("  Autor: %s (%s)\n", commit.Author.Name, commit.Author.Email))
		result.WriteString(fmt.Sprintf("  Data: %s\n", commit.Author.Date))
		result.WriteString("\n")
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: CallToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": result.String(),
				},
			},
		},
	}
}

func (s *MCPServer) handleGetContent(ctx context.Context, msg MCPMessage, params CallToolParams) MCPMessage {
	owner, _ := params.Arguments["owner"].(string)
	repo, _ := params.Arguments["repo"].(string)
	path, _ := params.Arguments["path"].(string)

	content, err := s.github.GetContent(ctx, owner, repo, path)
	if err != nil {
		return MCPMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Conteúdo de %s/%s/%s:\n\n", owner, repo, path))
	result.WriteString(fmt.Sprintf("Tipo: %s\n", content.Type))
	result.WriteString(fmt.Sprintf("Tamanho: %d bytes\n", content.Size))
	result.WriteString(fmt.Sprintf("URL: %s\n", content.HTMLURL))

	if content.Content != "" {
		result.WriteString("\nConteúdo:\n")
		result.WriteString(content.Content)
	}

	return MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: CallToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": result.String(),
				},
			},
		},
	}
}

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN não definido")
	}

	server := NewMCPServer(token)
	ctx := context.Background()

	log.Println("Servidor MCP GitHub iniciado")
	log.Println("Aguardando mensagens via stdin...")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg MCPMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("Erro ao parsear JSON: %v", err)
			continue
		}

		response := server.HandleMessage(ctx, msg)
		responseJSON, _ := json.Marshal(response)
		fmt.Println(string(responseJSON))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Erro ao ler stdin: %v", err)
	}
}