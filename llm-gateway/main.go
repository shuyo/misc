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

const maxBodyBytes = 4 * 1024 * 1024

type Config struct {
	Listen string  `json:"listen"`
	Routes []Route `json:"routes"`
}

type Route struct {
	BaseURL   string   `json:"base_url"`
	APIKeyEnv string   `json:"api_key_env"`
	Models    []string `json:"models"`
var proxyPostPaths = map[string]struct{}{
	"/v1/chat/completions": {},
	"/v1/completions":      {},
	"/v1/embeddings":       {},
	"/v1/rerank":           {},
	"/v2/rerank":           {},
	mux.HandleFunc("/v1/chat/completions", g.handleOpenAIProxy)
	mux.HandleFunc("/v1/completions", g.handleOpenAIProxy)
	mux.HandleFunc("/v1/embeddings", g.handleOpenAIProxy)
	mux.HandleFunc("/v1/rerank", g.handleOpenAIProxy)
	mux.HandleFunc("/v2/rerank", g.handleOpenAIProxy)
	if err != nil {
func (g *Gateway) handleOpenAIProxy(w http.ResponseWriter, r *http.Request) {
	}
	if _, ok := proxyPostPaths[r.URL.Path]; !ok {
		http.NotFound(w, r)
		return
	}

	rt, err := g.matchRoute(reqModel)
	upstreamURL := strings.TrimRight(rt.BaseURL, "/") + r.URL.Path
	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(body))
func (g *Gateway) matchRoute(model string) (Route, error) {
				return rt, nil

	return Route{}, fmt.Errorf("unsupported model: %s", model)
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
	for i, rt := range cfg.Routes {
		if strings.TrimSpace(rt.BaseURL) == "" {
			return Config{}, fmt.Errorf("routes[%d].base_url is required", i)
		}
		if len(rt.Models) == 0 {
			return Config{}, fmt.Errorf("routes[%d].models is required", i)
		}
	}
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
		proxyStream(w, upResp, rt.Name)
		return
	}
	proxyBuffer(w, upResp, rt.Name)
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

	modelRaw, ok := raw["model"]

	if err := json.Unmarshal(modelRaw, &model); err != nil || strings.TrimSpace(model) == "" {

	if streamRaw, ok := raw["stream"]; ok {
		if err := json.Unmarshal(streamRaw, &stream); err != nil {

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

func proxyStream(w http.ResponseWriter, upResp *http.Response, routeName string) {
	copyHeaders(w.Header(), upResp.Header)
	if routeName != "" {
		w.Header().Set("X-LLM-Gateway-Route", routeName)
	}
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

func proxyBuffer(w http.ResponseWriter, upResp *http.Response, routeName string) {
	copyHeaders(w.Header(), upResp.Header)
	if routeName != "" {
		w.Header().Set("X-LLM-Gateway-Route", routeName)
	}
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
