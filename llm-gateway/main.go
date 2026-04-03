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
	"net/url"
	"os"
	"strings"
	"time"
)

const maxBodyBytes = 4 * 1024 * 1024

type Config struct {
	Listen string  `json:"listen"`
	Routes []Route `json:"routes"`
}

type Route struct {
	Name        string   `json:"name"`
	BaseURL     string   `json:"base_url"`
	APIKeyEnv   string   `json:"api_key_env"`
	StripPrefix string   `json:"strip_prefix"`
	Models      []string `json:"models"`
}

type Gateway struct {
	cfg    Config
	client *http.Client
}

type endpointPolicy struct {
	Path           string
	AllowExtraBody bool
}

var endpointPolicies = map[string]endpointPolicy{
	"chat_completions": {
		Path:           "/v1/chat/completions",
		AllowExtraBody: true,
	},
	"completions": {
		Path: "/v1/completions",
	},
	"embeddings": {
		Path: "/v1/embeddings",
	},
	"rerank": {
		Path: "/v1/rerank",
	},
	"rerank_v2": {
		Path: "/v2/rerank",
	},
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
	mux.HandleFunc("/v1/chat/completions", g.wrapProxy("chat_completions"))
	mux.HandleFunc("/v1/completions", g.wrapProxy("completions"))
	mux.HandleFunc("/v1/embeddings", g.wrapProxy("embeddings"))
	mux.HandleFunc("/v1/rerank", g.wrapProxy("rerank"))
	mux.HandleFunc("/v2/rerank", g.wrapProxy("rerank_v2"))

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

func (g *Gateway) wrapProxy(policyName string) http.HandlerFunc {
	policy, ok := endpointPolicies[policyName]
	if !ok {
		panic("invalid endpoint policy")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		g.handleOpenAIProxy(w, r, policy)
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

func (g *Gateway) handleOpenAIProxy(w http.ResponseWriter, r *http.Request, policy endpointPolicy) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
		return
	}

	reqModel, stream, err := extractModelAndStream(body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	rt, upstreamModel, err := g.matchRoute(reqModel)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	bodyForUpstream := body
	if policy.AllowExtraBody {
		bodyForUpstream, err = expandExtraBody(body)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
			return
		}
	}

	patchedBody, err := replaceModel(bodyForUpstream, upstreamModel)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "invalid model payload")
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	upstreamURL, err := buildUpstreamURL(rt.BaseURL, policy.Path)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(patchedBody))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "server_error", "failed to build upstream request")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	applyAuthorizationHeader(upReq, r, rt)

	upResp, err := g.client.Do(upReq)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "upstream_error", fmt.Sprintf("failed to call upstream: %v", err))
		return
	}
	defer upResp.Body.Close()

	if stream {
		proxyStream(w, upResp)
		return
	}
	proxyBuffer(w, upResp)
}

func extractModelAndStream(body []byte) (string, bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", false, fmt.Errorf("invalid JSON")
	}
	v, ok := raw["model"]
	if !ok {
		return "", false, fmt.Errorf("model is required")
	}
	var model string
	if err := json.Unmarshal(v, &model); err != nil || strings.TrimSpace(model) == "" {
		return "", false, fmt.Errorf("model must be a non-empty string")
	}
	var stream bool
	if sv, ok := raw["stream"]; ok {
		if err := json.Unmarshal(sv, &stream); err != nil {
			return "", false, fmt.Errorf("stream must be boolean")
		}
	}
	return model, stream, nil
}

func (g *Gateway) matchRoute(model string) (Route, string, error) {
	for _, rt := range g.cfg.Routes {
		for _, declared := range rt.Models {
			if model == declared {
				return rt, normalizeModelForUpstream(rt, model), nil
			}
			if model == normalizeModelForUpstream(rt, declared) {
				return rt, normalizeModelForUpstream(rt, model), nil
			}
		}
	}

	return Route{}, "", fmt.Errorf("unsupported model: %s", model)
}

func normalizeModelForUpstream(rt Route, model string) string {
	if rt.StripPrefix != "" && strings.HasPrefix(model, rt.StripPrefix) {
		return strings.TrimPrefix(model, rt.StripPrefix)
	}
	return model
}

func expandExtraBody(body []byte) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}
	extraRaw, ok := raw["extra_body"]
	if !ok {
		return body, nil
	}
	var extra map[string]json.RawMessage
	if err := json.Unmarshal(extraRaw, &extra); err != nil {
		return nil, fmt.Errorf("extra_body must be a JSON object")
	}
	for k, v := range extra {
		if _, exists := raw[k]; exists {
			continue
		}
		raw[k] = v
	}
	delete(raw, "extra_body")
	return json.Marshal(raw)
}

func buildUpstreamURL(baseURL, path string) (string, error) {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return "", fmt.Errorf("route base_url is required")
	}
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid route base_url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid route base_url: %s", baseURL)
	}
	u.Path = ""
	return strings.TrimRight(u.String(), "/") + path, nil
}

func applyAuthorizationHeader(upReq *http.Request, inbound *http.Request, rt Route) {
	if rt.APIKeyEnv != "" {
		if key := strings.TrimSpace(os.Getenv(rt.APIKeyEnv)); key != "" {
			upReq.Header.Set("Authorization", "Bearer "+key)
			return
		}
	}
	if auth := strings.TrimSpace(inbound.Header.Get("Authorization")); auth != "" {
		upReq.Header.Set("Authorization", auth)
	}
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
