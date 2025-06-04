# TRICOLOUR Game Tools

このリポジトリはボードゲーム「TRICOLOUR」を扱う小さな Python スクリプト集です。
`tricolour/` ディレクトリの下に 3 つのスクリプトがあり、Python 3 で動作します。

## ディレクトリ構成

```
tricolour/
├── analyze.py   # ゲームログ解析ツール
├── gametree.py  # ゲーム木生成ツール
└── tricolore.py # 盤面・AI 実装メインモジュール
```

## 各スクリプト概要

- **tricolore.py**
  - ボードを長さ 57 の一次元配列で表現し、外周に番兵セル (GUARD) を置く実装です。
  - `directions` で周囲 8 方向のオフセットを定義し、石を置く処理 (`putstone`, `putstoneW`) や局面評価 (`score`) を提供します。
  - ランダム手、貪欲法、簡易 MinMax など複数の AI クラスが組み込まれており、`match` で対局、`statistics` で成績集計が可能です。

- **gametree.py**
  - `tricolore.py` の関数を利用してゲーム木を構築するツールです。コマンドライン引数で探索の最大深さを指定できます。

- **analyze.py**
  - ログを読み取り、正規表現で手をパースしながらゲーム木を辿り、勝敗を集計します。シンプルな再帰解析 (`path1`, `path2`) とツリー出力 (`printtree`) を行います。

## 知っておくと良いポイント

- **盤面表現**
  - 番兵セルを含む 1 次元配列を使っているため、座標変換には `pos2tuple` / `tuple2pos` を確認すると理解しやすいでしょう。
- **白石の特別ルール**
  - `putstoneW` など "白石" 用の処理があり、通常の石とは反転の条件が異なります。
- **AI クラス**
  - `RandomPlayer`, `RandomPlayer2`, `RandomPlayer3`, `Greedy`, `MinMax` などサンプル実装が用意されています。挙動を確認しながらカスタマイズすると全体像を掴みやすくなります。
  - **実装言語**
  - Python 3 で書かれています。

## 使い方

```
# 例: 対局を実行
python3 tricolour/tricolore.py

# 例: ゲーム木を深さ3で生成
python3 tricolour/gametree.py 3

# 例: ログ解析
python3 tricolour/analyze.py LOGFILE
```

## 次のステップ

1. まず `tricolore.py` を読んで基本的な盤面処理や AI の仕組みを理解しましょう。
2. `gametree.py` を実行して探索がどのように行われているか試してみてください。
3. AI クラスを拡張したり新しい戦略を追加して、自動対局を行ってみましょう。
4. Python 3 への移植や、機能追加を通じてコードベースに慣れると良いでしょう。

ライセンスは MIT です。お気軽にご活用ください。
