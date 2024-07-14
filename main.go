package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RequestInfo struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

var (
	webhooks   = make(map[string][]RequestInfo)
	webhooksMu sync.RWMutex
	clients    = make(map[string]map[chan string]bool)
	clientsMu  sync.Mutex
	tmpl       *template.Template
)

func main() {
	var err error
	tmpl, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatalf("Erro ao carregar o template: %v", err)
	}

	http.HandleFunc("/webhook/", webhookHandler)

	fmt.Println("Servidor iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Endpoint do webhook não fornecido", http.StatusBadRequest)
		return
	}

	endpoint := parts[2]

	switch r.Method {
	case http.MethodGet:
		handleMonitorPage(w, endpoint)
	case http.MethodPost:
		handleWebhookPost(w, r, endpoint)
	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

func handleMonitorPage(w http.ResponseWriter, endpoint string) {
	tmpl.Execute(w, map[string]string{"Endpoint": endpoint})
}

func handleWebhookPost(w http.ResponseWriter, r *http.Request, endpoint string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erro ao ler o corpo do request", http.StatusInternalServerError)
		return
	}

	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	requestInfo := RequestInfo{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Method:    r.Method,
		Headers:   headers,
		Body:      string(body),
	}

	webhooksMu.Lock()
	webhooks[endpoint] = append(webhooks[endpoint], requestInfo)
	webhooksMu.Unlock()

	// Transmita o novo webhook
	go broadcastWebhook(endpoint, requestInfo)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Webhook recebido com sucesso. Endpoint: %s, ID: %s", endpoint, requestInfo.ID)
}

func broadcastWebhook(endpoint string, webhook RequestInfo) {
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		log.Printf("Erro ao converter webhook para JSON: %v", err)
		return
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	if clients[endpoint] == nil {
		return
	}

	for clientChan := range clients[endpoint] {
		clientChan <- string(jsonData)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request, endpoint string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE não suportado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan string)

	clientsMu.Lock()
	if clients[endpoint] == nil {
		clients[endpoint] = make(map[chan string]bool)
	}
	clients[endpoint][clientChan] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients[endpoint], clientChan)
		if len(clients[endpoint]) == 0 {
			delete(clients, endpoint)
		}
		clientsMu.Unlock()
	}()

	// Envie webhooks existentes para o novo cliente
	webhooksMu.RLock()
	for _, webhook := range webhooks[endpoint] {
		jsonData, _ := json.Marshal(webhook)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
	}
	webhooksMu.RUnlock()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func init() {
	http.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "Endpoint não fornecido", http.StatusBadRequest)
			return
		}
		endpoint := parts[2]
		sseHandler(w, r, endpoint)
	})
}
