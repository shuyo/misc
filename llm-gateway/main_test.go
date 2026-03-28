package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestExtractModelAndStream(t *testing.T) {
	model, stream, err := extractModelAndStream([]byte(`{"model":"openai/gpt-4.1-mini","stream":true}`))
	if err != nil {
		t.Fatalf("extractModelAndStream: %v", err)
	}
	if model != "openai/gpt-4.1-mini" {
		t.Fatalf("unexpected model: %s", model)
	}
	if !stream {
		t.Fatalf("expected stream true")
	}
}

func TestMatchRoute(t *testing.T) {
	g := &Gateway{cfg: Config{Routes: []Route{{
		Name:          "openai",
		BaseURL:       "https://api.openai.com",
		ModelPrefixes: []string{"openai/", "gpt-"},
		StripPrefix:   "openai/",
	}}}}

	rt, upstream, err := g.matchRoute("openai/gpt-4.1-mini")
	if err != nil {
		t.Fatalf("match route: %v", err)
	}
	if rt.Name != "openai" {
		t.Fatalf("unexpected route: %s", rt.Name)
	}
	if upstream != "gpt-4.1-mini" {
		t.Fatalf("unexpected upstream model: %s", upstream)
	}
}

func TestExpandExtraBody(t *testing.T) {
	merged, err := expandExtraBody([]byte(`{"model":"openai/gpt-4.1-mini","messages":[],"extra_body":{"foo":1,"bar":"x"}}`))
	if err != nil {
		t.Fatalf("expand extra_body: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(merged, &raw); err != nil {
		t.Fatalf("unmarshal merged body: %v", err)
	}
	if _, ok := raw["foo"]; !ok {
		t.Fatalf("expected merged field foo")
	}
	if _, ok := raw["bar"]; !ok {
		t.Fatalf("expected merged field bar")
	}
	if _, ok := raw["extra_body"]; ok {
		t.Fatalf("extra_body should be removed after merge")
	}
}

func TestMatchRouteByExactDeclaredModel(t *testing.T) {
	g := &Gateway{cfg: Config{Routes: []Route{{
		Name:          "vllm",
		BaseURL:       "http://127.0.0.1:8001",
		ModelPrefixes: []string{"vllm/"},
		Models:        []string{"Qwen3-0.6B"},
	}}}}

	rt, upstream, err := g.matchRoute("Qwen3-0.6B")
	if err != nil {
		t.Fatalf("match route: %v", err)
	}
	if rt.Name != "vllm" {
		t.Fatalf("unexpected route: %s", rt.Name)
	}
	if upstream != "Qwen3-0.6B" {
		t.Fatalf("unexpected upstream model: %s", upstream)
	}
}

func TestMatchRouteSingleRouteFallback(t *testing.T) {
	g := &Gateway{cfg: Config{Routes: []Route{{
		Name:    "vllm",
		BaseURL: "http://127.0.0.1:8001",
	}}}}

	rt, upstream, err := g.matchRoute("Any-Model-Name")
	if err != nil {
		t.Fatalf("match route fallback: %v", err)
	}
	if rt.Name != "vllm" {
		t.Fatalf("unexpected route: %s", rt.Name)
	}
	if upstream != "Any-Model-Name" {
		t.Fatalf("unexpected upstream model: %s", upstream)
	}
}

func TestBuildUpstreamURLWithoutScheme(t *testing.T) {
	u, err := buildUpstreamURL("127.0.0.1:8000", "/v1/chat/completions")
	if err != nil {
		t.Fatalf("buildUpstreamURL: %v", err)
	}
	if u != "http://127.0.0.1:8000/v1/chat/completions" {
		t.Fatalf("unexpected upstream url: %s", u)
	}
}

func TestApplyAuthorizationHeaderFallbackToInbound(t *testing.T) {
	upReq, _ := http.NewRequest("POST", "http://example.com", nil)
	inReq, _ := http.NewRequest("POST", "http://localhost", nil)
	inReq.Header.Set("Authorization", "Bearer inbound-token")
	applyAuthorizationHeader(upReq, inReq, Route{})
	if got := upReq.Header.Get("Authorization"); got != "Bearer inbound-token" {
		t.Fatalf("unexpected authorization header: %s", got)
	}
}

func TestRerankV2Policy(t *testing.T) {
	p, ok := endpointPolicies["rerank_v2"]
	if !ok {
		t.Fatalf("rerank_v2 policy not found")
	}
	if p.Path != "/v2/rerank" {
		t.Fatalf("unexpected rerank_v2 path: %s", p.Path)
	}
}
