package main

import "testing"

func TestExtractStream(t *testing.T) {
	raw, err := validateSupportedFields([]byte(`{"model":"openai/gpt-4.1-mini","stream":true}`), endpointPolicies["chat_completions"].AllowedFields, "x")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	stream, err := extractStream(raw)
	if err != nil {
		t.Fatalf("extract stream: %v", err)
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

func TestValidateUnsupportedField(t *testing.T) {
	_, err := validateSupportedFields([]byte(`{"model":"x","foo":1}`), endpointPolicies["embeddings"].AllowedFields, "unsupported field for embeddings")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestExpandExtraBody(t *testing.T) {
	merged, err := expandExtraBody([]byte(`{"model":"openai/gpt-4.1-mini","messages":[],"extra_body":{"foo":1,"bar":"x"}}`))
	if err != nil {
		t.Fatalf("expand extra_body: %v", err)
	}
	raw, err := validateSupportedFields(merged, setOf("model", "messages", "foo", "bar"), "x")
	if err != nil {
		t.Fatalf("validate merged body: %v", err)
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
