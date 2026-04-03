# llm-gateway

OpenAI API 互換の最小依存ゲートウェイです。複数上流 LLM サービスを 1 エンドポイントで扱えます。

## 特徴
- Go 標準ライブラリのみ（外部パッケージ依存なし）
- `routes[].models` の**完全一致のみ**でルーティング（prefix マッチやフォールバックなし）
- `base_url` は `http://` / `https://` 省略時に自動で `http://` を補完
- `api_key_env` が未設定または空の場合、受信した `Authorization` ヘッダをそのまま上流へ転送
- ストリーミング時は upstream を pass-through し、`Flush()` で逐次送信
- クライアント切断時は `context` で upstream を即 cancel
- 基本は上流へそのまま透過するシンプルな proxy（`chat/completions` の `extra_body` は展開して転送）
`routes[]` で使用するキーは `base_url`, `api_key_env`, `strip_prefix`, `models` のみです（`name` は使いません）。

- どの route が使われたかをレスポンスヘッダ `X-LLM-Gateway-Route` で返す（`routes[].name` を使用）

## 対応 API
- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/completions`
- `POST /v1/embeddings`
- `POST /v1/rerank`
- `POST /v2/rerank`

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

curl -s http://localhost:8080/v2/rerank \
  -H 'Content-Type: application/json' \
  -d '{
    "model":"openai/rerank-v1",
    "query":"supply chain",
    "documents":["doc1","doc2"],
    "top_n":1
  }'
```
