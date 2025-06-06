# TRICOLOUR ルール概要

**TRICOLOUR** は 6×6 のリバーシ風ボードゲームです。プレイヤーは赤と青の 2 色に分かれますが、各色には「白石」状態が存在します。白石は盤面には置かれるものの得点計算には含まれません。

## 初期配置
- 盤面は一次元配列で表され、番兵セル (GUARD) を含め長さ 57 です。【F:README.md†L17-L19】
- 中央 4 マスに赤 2 個、青 2 個を配置した状態から開始します。【F:tricolour/tricolore.py†L36-L44】

## 手番の流れ
1. 自分の色の石を置くか、白石を置くか選択します。
2. 石を置けるマスは `availableplaces` で判定され、通常石用と白石用の候補が別々に返ります。【F:tricolour/tricolore.py†L78-L87】
3. 選択したマスに置いた後、裏返せる石を探索して反転させます。通常石は `putstone`、白石は `putstoneW` が処理します。

## 石の反転ルール
- **通常石 (`putstone`)**
  - 隣接マスから一直線上に、自分の色で挟める石がある場合に置けます。【F:tricolour/tricolore.py†L54-L64】
  - 挟んだ石は `^= 2` で白⇔色を入れ替えます。【F:tricolour/tricolore.py†L95-L107】
- **白石 (`putstoneW`)**
  - 最初に隣接する石が赤または青で、直線上の端に白石がある場合に置けます。【F:tricolour/tricolore.py†L66-L76】
  - 挟んだ石は同じく `^= 2` で白⇔色を入れ替えます。【F:tricolour/tricolore.py†L110-L129】

## 勝敗
- 盤面に空きマスがなくなる、両者が連続パスする、または赤か青の石が 0 になった時点で終了します。【F:tricolour/tricolore.py†L318-L325】
- `score` は赤と青の「色付き」石のみを数え、白石は得点になりません。【F:tricolour/tricolore.py†L131-L138】
- 得点が多い方の色が勝者となります。

