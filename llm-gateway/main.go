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

type endpointPolicy struct {
	Path            string
	AllowStream     bool
	AllowExtraBody  bool
	RequireModel    bool
	AllowedFields   map[string]struct{}
	UnsupportedText string
}

var endpointPolicies = map[string]endpointPolicy{
	"chat_completions": {
		Path:           "/v1/chat/completions",
		AllowStream:    true,
		AllowExtraBody: true,
		RequireModel:   true,
		AllowedFields: setOf(
			"model", "messages", "stream", "temperature", "top_p", "max_tokens", "presence_penalty", "frequency_penalty", "stop", "n", "user", "extra_body",
		),
		UnsupportedText: "unsupported field for chat/completions",
	},
	"completions": {
		Path:         "/v1/completions",
		AllowStream:  true,
		RequireModel: true,
		AllowedFields: setOf(
			"model", "prompt", "suffix", "max_tokens", "temperature", "top_p", "n", "stream", "logprobs", "echo", "stop", "presence_penalty", "frequency_penalty", "best_of", "logit_bias", "user",
		),
		UnsupportedText: "unsupported field for completions",
	},
	"embeddings": {
		Path:         "/v1/embeddings",
		AllowStream:  false,
		RequireModel: true,
		AllowedFields: setOf(
			"model", "input", "encoding_format", "dimensions", "user",
		),
		UnsupportedText: "unsupported field for embeddings",
	},
	"rerank": {
		Path:         "/v1/rerank",
		AllowStream:  false,
		RequireModel: true,
		AllowedFields: setOf(
			"model", "query", "documents", "top_n", "return_documents", "max_chunks_per_doc", "user",
		),
		UnsupportedText: "unsupported field for rerank",
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

	raw, err := validateSupportedFields(body, policy.AllowedFields, policy.UnsupportedText)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	reqModel, err := extractModel(raw, policy.RequireModel)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	rt, upstreamModel, err := g.matchRoute(reqModel)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	stream, err := extractStream(raw)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	if stream && !policy.AllowStream {
		writeErr(w, http.StatusBadRequest, "invalid_request_error", "stream is not supported for this endpoint")
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

func validateSupportedFields(body []byte, allowed map[string]struct{}, messagePrefix string) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}
	for k := range raw {
		if _, ok := allowed[k]; !ok {
			return nil, fmt.Errorf("%s: %s", messagePrefix, k)
		}
	}
	return raw, nil
}

func extractModel(raw map[string]json.RawMessage, required bool) (string, error) {
	if !required {
		return "", nil
	}
	v, ok := raw["model"]
	if !ok {
		return "", fmt.Errorf("model is required")
	}
	var model string
	if err := json.Unmarshal(v, &model); err != nil || strings.TrimSpace(model) == "" {
		return "", fmt.Errorf("model must be a non-empty string")
	}
	return model, nil
}

func extractStream(raw map[string]json.RawMessage) (bool, error) {
	v, ok := raw["stream"]
	if !ok {
		return false, nil
	}
	var stream bool
	if err := json.Unmarshal(v, &stream); err != nil {
		return false, fmt.Errorf("stream must be boolean")
	}
	return stream, nil
}

func (g *Gateway) matchRoute(model string) (Route, string, error) {
	for _, rt := range g.cfg.Routes {
		for _, p := range rt.ModelPrefixes {
			if strings.HasPrefix(model, p) {
				return rt, normalizeModelForUpstream(rt, model), nil
			}
		}
	}

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

	if len(g.cfg.Routes) == 1 {
		rt := g.cfg.Routes[0]
		return rt, normalizeModelForUpstream(rt, model), nil
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

func setOf(items ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, it := range items {
		m[it] = struct{}{}
	}
	return m
}
