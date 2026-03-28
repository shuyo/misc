# llm-gateway

OpenAI API 互換の最小依存ゲートウェイです。`POST /v1/chat/completions` と `GET /v1/models` のみ実装しています。

## 特徴
- Go 標準ライブラリのみ（外部パッケージ依存なし）
- `model` prefix ルーティングで複数上流サービスを 1 エンドポイントに集約
- ストリーミング時は upstream を pass-through し、`Flush()` で逐次送信
- クライアント切断時は `context` で upstream を即 cancel
- 未対応パラメータは 400 で明示

## 使い方
```bash
cp config.example.json config.json
export OPENAI_API_KEY=...
export ANTHROPIC_API_KEY=...
go run . -config ./config.json
```

## API
- `GET /v1/models`
- `POST /v1/chat/completions`

`/v1/chat/completions` は OpenAI 互換 JSON を受け取り、`model` に応じて上流へ転送します。

### 例
```bash
curl -s http://localhost:8080/v1/models

curl -N http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "openai/gpt-4.1-mini",
    "messages": [{"role":"user","content":"hello"}],
    "stream": true
  }'
```
