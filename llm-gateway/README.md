# llm-gateway

OpenAI API 互換の最小依存ゲートウェイです。複数上流 LLM サービスを 1 エンドポイントで扱えます。

## 特徴
- Go 標準ライブラリのみ（外部パッケージ依存なし）
- `model` prefix ルーティングで複数上流サービスを集約
- ストリーミング時は upstream を pass-through し、`Flush()` で逐次送信
- クライアント切断時は `context` で upstream を即 cancel
- 未対応パラメータは 400 で明示（`chat/completions` の `extra_body` は例外的に上流へ透過）

## 対応 API
- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/completions`
- `POST /v1/embeddings`
- `POST /v1/rerank`

## 使い方
```bash
cp config.example.json config.json
export OPENAI_API_KEY=...
export ANTHROPIC_API_KEY=...
go run . -config ./config.json
```


## Docker (multi-stage build)
```bash
docker build -t llm-gateway:local .
docker run --rm -p 8080:8080 \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -v $(pwd)/config.json:/app/config.json:ro \
  llm-gateway:local
```

## 例
```bash
curl -s http://localhost:8080/v1/models

curl -N http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "openai/gpt-4.1-mini",
    "messages": [{"role":"user","content":"hello"}],
    "stream": true
  }'

# extra_body を使ってプロバイダ固有パラメータを透過
curl -N http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "openai/gpt-4.1-mini",
    "messages": [{"role":"user","content":"hello"}],
    "extra_body": {"reasoning": {"effort": "medium"}}
  }'

curl -s http://localhost:8080/v1/embeddings \
  -H 'Content-Type: application/json' \
  -d '{"model":"openai/text-embedding-3-large","input":"hello"}'

curl -s http://localhost:8080/v1/rerank \
  -H 'Content-Type: application/json' \
  -d '{
    "model":"openai/rerank-v1",
    "query":"supply chain",
    "documents":["doc1","doc2"],
    "top_n":1
  }'
```
