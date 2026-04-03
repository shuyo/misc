package main

import (
		BaseURL: "https://api.openai.com",
		Models:  []string{"openai/gpt-4.1-mini"},
	_, err := g.matchRoute("openai/gpt-4.1-mini")

	_, err := g.matchRoute("Qwen3-0.6B")
	if err != nil {
		t.Fatalf("match route: %v", err)
	}
	_, err := g.matchRoute("Any-Model-Name")
func TestRerankV2PathAllowed(t *testing.T) {
	_, ok := proxyPostPaths["/v2/rerank"]
		t.Fatalf("/v2/rerank should be in allowed proxy paths")
	if _, ok := raw["bar"]; !ok {
		t.Fatalf("expected merged field bar")
	}
	if _, ok := raw["extra_body"]; ok {
		t.Fatalf("extra_body should be removed after merge")
	}
}

func TestMatchRouteByExactDeclaredModel(t *testing.T) {
	g := &Gateway{cfg: Config{Routes: []Route{{
		Name:    "vllm",
		BaseURL: "http://127.0.0.1:8001",
		Models:  []string{"Qwen3-0.6B"},
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

func TestMatchRouteWithoutExactMatchReturnsError(t *testing.T) {
	g := &Gateway{cfg: Config{Routes: []Route{{
		Name:    "vllm",
		BaseURL: "http://127.0.0.1:8001",
		Models:  []string{"Qwen3-0.6B"},
	}}}}

	_, _, err := g.matchRoute("Any-Model-Name")
	if err == nil {
		t.Fatalf("expected route matching error")
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

func TestProxyBufferSetsRouteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	upResp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"error":"x"}`)),
	}
	proxyBuffer(rec, upResp, "vllm")
	if got := rec.Header().Get("X-LLM-Gateway-Route"); got != "vllm" {
		t.Fatalf("unexpected route header: %s", got)
	}
}
