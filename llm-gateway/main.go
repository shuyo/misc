package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Listen string  `json:"listen"`
	Routes []Route `json:"routes"`
}

type Route struct {
	Name          string   `json:"name"`
	BaseURL       string   `json:"base_url"`
	APIKeyEnv     string   `json:"api_key_env"`
	ModelPrefixes []string `json:"model_prefixes"`
	StripPrefix   string   `json:"strip_prefix"`
	Models        []string `json:"models"`
}

type Gateway struct {
	cfg    Config
	client *http.Client
}

func main() {
	configPath := flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	g := &Gateway{
		cfg: cfg,
		client: &http.Client{
			Timeout: 0,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/models", g.handleModels)
	mux.HandleFunc("/v1/chat/completions", g.handleChatCompletions)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("listening on %s", cfg.Listen)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}
}

func loadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if len(cfg.Routes) == 0 {
		return Config{}, fmt.Errorf("routes is required")
	}
	return cfg, nil
}

func (g *Gateway) handleModels(w http.ResponseWriter, _ *http.Request) {
	type modelObj struct {
		ID     string `json:"id"`
		Object string `json:"object"`
	}
	resp := struct {
		Object string     `json:"object"`
		Data   []modelObj `json:"data"`
	}{
		Object: "list",
		Data:   make([]modelObj, 0, 8),
	}

	seen := map[string]struct{}{}
	for _, rt := range g.cfg.Routes {
		for _, m := range rt.Models {
			if _, ok := seen[m]; ok {
				continue
			}
			seen[m] = struct{}{}
			resp.Data = append(resp.Data, modelObj{ID: m, Object: "model"})
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (g *Gateway) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4*1024*1024))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	if err := validateSupportedFields(body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON")
		return
	}
	if req.Model == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}

	rt, upstreamModel, err := g.matchRoute(req.Model)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	patchedBody, err := replaceModel(body, upstreamModel)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "invalid model payload")
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	upstreamURL := strings.TrimRight(rt.BaseURL, "/") + "/v1/chat/completions"
	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(patchedBody))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "server_error", "failed to build upstream request")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	if rt.APIKeyEnv != "" {
		if key := strings.TrimSpace(os.Getenv(rt.APIKeyEnv)); key != "" {
			upReq.Header.Set("Authorization", "Bearer "+key)
		}
	}

	upResp, err := g.client.Do(upReq)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "upstream_error", "failed to call upstream")
		return
	}
	defer upResp.Body.Close()

	if req.Stream {
		proxyStream(w, upResp)
		return
	}
	proxyBuffer(w, upResp)
}

func validateSupportedFields(body []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("invalid JSON")
	}
	allowed := map[string]struct{}{
		"model":             {},
		"messages":          {},
		"stream":            {},
		"temperature":       {},
		"top_p":             {},
		"max_tokens":        {},
		"presence_penalty":  {},
		"frequency_penalty": {},
		"stop":              {},
		"n":                 {},
		"user":              {},
	}
	for k := range raw {
		if _, ok := allowed[k]; !ok {
			return fmt.Errorf("unsupported field: %s", k)
		}
	}
	return nil
}

func (g *Gateway) matchRoute(model string) (Route, string, error) {
	for _, rt := range g.cfg.Routes {
		for _, p := range rt.ModelPrefixes {
			if strings.HasPrefix(model, p) {
				if rt.StripPrefix != "" && strings.HasPrefix(model, rt.StripPrefix) {
					return rt, strings.TrimPrefix(model, rt.StripPrefix), nil
				}
				return rt, model, nil
			}
		}
	}
	return Route{}, "", fmt.Errorf("unsupported model: %s", model)
}

func replaceModel(body []byte, model string) ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	raw["model"] = model
	return json.Marshal(raw)
}

func proxyStream(w http.ResponseWriter, upResp *http.Response) {
	copyHeaders(w.Header(), upResp.Header)
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(upResp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	buf := make([]byte, 8*1024)
	for {
		n, err := upResp.Body.Read(buf)
		if n > 0 {
			_, _ = w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}
	}
}

func proxyBuffer(w http.ResponseWriter, upResp *http.Response) {
	copyHeaders(w.Header(), upResp.Header)
	w.WriteHeader(upResp.StatusCode)
	_, _ = io.Copy(w, upResp.Body)
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func writeErr(w http.ResponseWriter, code int, typ, message string) {
	writeJSON(w, code, map[string]any{
		"error": map[string]any{
			"type":    typ,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
