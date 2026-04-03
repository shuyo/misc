# llm-proxy 簡素化提案（実装前）

このドキュメントは**コード変更前の提案のみ**をまとめたものです。

## 現状で長くなっている主な要因
- エンドポイントごとの分岐 (`/v1/...`, `/v2/...`) を手動で並べている
- `chat/completions` の `extra_body` 展開など、部分的な前処理がある
- ルーティング・認証・URL 正規化・streaming 転送を 1 ファイルで保持している

## まず効く「低リスク」簡素化候補

### 1) `endpointPolicies` をやめて「パスそのまま転送」へ寄せる
- 現状: policy 名 (`chat_completions` など) を内部 map で path に変換
- 提案: 登録するハンドラで path を直接渡す（または `r.URL.Path` をそのまま使用）
- 効果: policy 定義と lookup の層を削減できる
- 影響: `extra_body` を chat だけ許可する条件分岐の置き場所を別途決める必要

### 2) `/v1/*` と `/v2/*` の個別 `HandleFunc` を減らす
- 現状: `/v1/chat/completions` などを個別登録
- 提案: 許可対象 path の set を 1 つ持ち、共通ハンドラ 1 本で処理
- 効果: ルート追加時の修正点が減る
- 影響: `models` endpoint (`GET`) と `POST` 系を分ける最低限の分岐は残る

### 3) `extractModelAndStream` + `replaceModel` を 1 回の JSON decode/encode に統合
- 現状: model/stream 抽出と model 更新で JSON を複数回処理
- 提案: 1 度 decode して model, stream を取り、必要なら更新して encode
- 効果: 関数数・JSON 変換回数が減る
- 影響: 関数責務が増えるので命名とテストを整理する必要

## 条件付きで有効な簡素化候補

### 4) `strip_prefix` を廃止
- 前提: upstream 側モデル名と公開モデル名を同一にする運用が可能
- 効果: `normalizeModelForUpstream` と関連テストを削減
- 影響: 既存クライアントの model 名が変わる可能性

### 5) `extra_body` 機能の廃止
- 前提: プロバイダ固有拡張を使わない
- 効果: `expandExtraBody` と chat 特例分岐を削減
- 影響: 一部 provider 拡張パラメータが渡せなくなる

### 6) `buildUpstreamURL` の柔軟性縮小
- 前提: `base_url` は必ず `http://` / `https://` 付きで運用統一
- 効果: 自動補完ロジック削減、設定ミスは起動時検証へ寄せられる
- 影響: 既存の「スキーム省略許容」がなくなる

## streaming を維持したまま簡素化する際の注意
- `proxyStream` の `Flush()` 呼び出しは残す
- レスポンスの status/header/body pass-through は維持する
- stream 途中でリトライしない

## 推奨する進め方（段階的）
1. 低リスク候補 1〜3 を適用
2. e2e で chat stream / non-stream を確認
3. 運用要件が許せば 4〜6 を順次検討

## どれから着手すべきか（提案）
- **第一候補:** 2) 共通ハンドラ化
- **第二候補:** 1) policy 廃止
- **第三候補:** 3) JSON 処理統合

この順なら、既存機能を保ちつつコード量削減効果が出やすいです。
