# ccplan 改善計画

## 概要

7つの改善項目について調査・計画をまとめる。作業量の見積もりは S/M/L で表記。

---

## 1. delete/approve → viewed フラグへの変更 [S] ✅ 実装済み

### 現状
- `d` キー: step に `[del]` バッジをトグル（`ActionDelete`）
- `a` キー: step に `[ok]` バッジをトグル（`ActionApprove`）
- submit 時、approve のみの場合は `StatusApproved` として扱われる

### 変更内容
- `ActionDelete` を完全に削除
- `ActionApprove` を削除し、代わりに `Viewed` フラグを導入
- `Viewed` はレビューコメントとは独立したフラグ（レビュー出力に含めない）
- `v` キーで viewed トグル、バッジは `[✓]`
- `d` キー、`a` キーのバインドを削除
- submit ロジック: approve 判定は「コメントが0件」の場合に変更

### 変更ファイル
| ファイル | 変更内容 |
|----------|----------|
| `internal/plan/model.go` | `ActionDelete`, `ActionApprove`, `ActionInsertAfter`, `ActionInsertBefore` を削除。`ActionModify` のみ残す |
| `internal/tui/steplist.go` | `StepListItem` に `Viewed bool` フィールド追加。`renderBadge` から delete/approve 分岐を削除、viewed バッジ `[✓]` を追加 |
| `internal/tui/app.go` | `handleLeftPaneKeys` から Delete/Approve ケースを削除、Viewed トグル (`v`) を追加。`submitReview` のapprove判定を「コメント0件」に変更 |
| `internal/tui/keymap.go` | `Delete`, `Approve` を削除、`Viewed` を追加 (`v` キー) |
| `internal/tui/comment.go` | action prefix パース部分を簡素化（modify のみ） |
| `internal/plan/review.go` | `FormatReview` の `[action]` 出力を変更 |
| `internal/tui/styles.go` | `DeleteBadge` → `ViewedBadge` に変更 (green 系で `[✓]`) |
| ステータスバー・ヘルプ | `d` `a` を `v` に変更 |

---

## 2. Conventional Comments 対応 [M] ✅ 実装済み

### 現状
- コメントの action type として `modify`, `delete`, `approve`, `insert-after`, `insert-before` がある
- テキストエリアに `modify:` などのプレフィックスを手入力すると action type が変わる
- Claude Code へのレビュー出力は `## S1: Title [modify]` のようなフォーマット

### Conventional Comments 仕様
フォーマット: `<label> [decorations]: <subject>`

標準ラベル:
| ラベル | 用途 |
|--------|------|
| `suggestion` | 改善提案 |
| `issue` | 具体的な問題の指摘 |
| `question` | 質問・確認 |
| `nitpick` | 些細な好みの問題 |
| `thought` | アイデア・参考情報 |
| `todo` | 必要な小さな変更 |
| `praise` | ポジティブな点 |
| `note` | 情報の強調 |
| `chore` | 受け入れ前に必要な作業 |

### 入力補助の提案

**案A: ラベルプレフィックスの自動挿入 + Tab サイクル (推奨)**

`c` を押してコメントエディタを開く際に、デフォルトで `suggestion: ` がプレフィックスとして挿入される。Tab キーでラベルをサイクルできる:

```
[Tab で切替: suggestion → issue → question → nitpick → todo → thought → note]

suggestion: |
```

- Tab を押すたびにラベル部分だけが切り替わる（本文はそのまま）
- 直接テキストを書き始めることも可能（ラベルなしコメント）
- decorations は省略（本ツールでは不要）

**案B: シンプルに modify のみ**

現行の `modify` を残し、ユーザーが手動で `suggestion:` などを書く。ツール側での入力補助なし。

### 推奨: 案A

理由:
- ラベルの存在を知らなくても使える（Tab を押さなければ suggestion がデフォルト）
- 実装がシンプル（Tab でプレフィックスを差し替えるだけ）
- decorations は省略してラベルのみに絞ることでUIを複雑にしない

### 変更内容 (案A の場合)

| ファイル | 変更内容 |
|----------|----------|
| `internal/plan/model.go` | `ActionType` をラベル型に変更。`ActionModify` → `LabelSuggestion` など。または ActionType 自体を string のままにして、値を conventional comment のラベルに変更 |
| `internal/tui/comment.go` | ラベル一覧の定義、Tab でサイクルするロジック、プレフィックス自動挿入。`Result()` でラベル + 本文の分離 |
| `internal/tui/app.go` | ModeComment での Tab キー処理をコメントエディタに委譲 |
| `internal/plan/review.go` | 出力フォーマットを `## S1: Title [suggestion]` に変更 |
| ステータスバー | コメントモード時に `tab` ラベル切替 の表示追加 |

### 注意
- 項目1 (delete/approve 削除) とセットで実装する
- ActionType を conventional comment ラベルに完全に置き換える

---

## 3. WezTerm ペインサイズの動的調整 [S] ✅ 実装済み

### 現状
- 固定: `--bottom --percent 80`

### 調査結果
`wezterm cli list --format json` で現在のペインサイズを取得可能:
```json
{
  "pane_id": 0,
  "size": { "rows": 59, "cols": 135 }
}
```

※ただし `is_active` フィールドの有無は WezTerm バージョンに依存する可能性がある。確実には `WEZTERM_PANE` 環境変数で現在のペインIDを取得し、マッチさせる。

### 変更内容

| ファイル | 変更内容 |
|----------|----------|
| `internal/pane/wezterm.go` | `SpawnAndWait` の冒頭で現在のペインサイズを取得。`cols > rows * 2` (横長) なら `--right --percent 50`、それ以外は `--bottom --percent 80` |

### ロジック

```go
func (w *WezTermSpawner) splitDirection() (string, string) {
    size, err := w.currentPaneSize()
    if err != nil {
        return "--bottom", "80" // fallback
    }
    if size.Cols > size.Rows*2 {
        return "--right", "50"
    }
    return "--bottom", "80"
}
```

`WEZTERM_PANE` 環境変数で現ペインIDを特定:
```go
func (w *WezTermSpawner) currentPaneSize() (*paneSize, error) {
    currentPaneID := os.Getenv("WEZTERM_PANE")
    // wezterm cli list --format json から currentPaneID に一致する pane の size を返す
}
```

---

## 4. macOS Terminal.app でのペインスポーン [M]

### 調査結果
AppleScript で Terminal.app を制御可能:
- `do script` でコマンド実行
- `busy` プロパティでコマンド完了を検出（ポーリング）
- ウィンドウのクローズも制御可能

### 制約
- Terminal.app が起動する（ユーザーの画面にウィンドウが出現する）
- 終了ステータスは直接取得できないが、ccplan は temp file IPC を使用しているため問題なし
- AppleScript の実行に macOS のアクセシビリティ許可が不要（Terminal.app 自体の操作は許可不要）

### 変更内容

| ファイル | 変更内容 |
|----------|----------|
| `internal/pane/terminal_darwin.go` | 新規。`TerminalAppSpawner` 実装。AppleScript で Terminal.app を制御 |
| `internal/pane/terminal_other.go` | 新規。非 darwin 用ビルドタグ。`Available()` が `false` を返す |
| `internal/pane/detect.go` | AutoDetect に `TerminalAppSpawner` を追加（WezTerm → tmux → Terminal.app → Direct） |

### 実装方針

```go
// terminal_darwin.go
type TerminalAppSpawner struct{}

func (t *TerminalAppSpawner) SpawnAndWait(cmd string, args []string) error {
    // 1. コマンドを組み立て
    // 2. AppleScript で Terminal.app を開いてコマンド実行
    // 3. busy プロパティをポーリングして完了を待つ
    // 4. ウィンドウを閉じる
}
```

### 注意点
- `DirectSpawner` は Claude Code のフック内で使うとターミナルが共有されて壊れるため、Terminal.app spawner は重要なフォールバック
- ビルドタグ (`//go:build darwin`) で macOS 限定にする

---

## 5. ステップ箱の幅拡張 [S]

### 現状
左ペインの幅: `width * 30 / 100` (width >= 80 の場合)

### 変更内容

| ファイル | 変更内容 |
|----------|----------|
| `internal/tui/app.go` | `leftWidth()` の計算を `width * 35 / 100` に変更。rightWidth も連動して調整される |

### 補足
- 30% → 35% で、120 カラムターミナルの場合 36 → 42 文字 (+6文字)
- 右ペインの幅は自動的に縮小するが、Markdown 表示には十分

---

## 6. 日本語入力 (IME) の問題 [保留]

### 調査結果
- **原因**: Bubble Tea v1 の制限。仮想カーソルを使用しているため、IME が実際のターミナルカーソル位置（左下）を参照する
- **変換前文字列の非表示**: ターミナルベース TUI 全般の制限
- **v2 での改善**: `tea.SetCursorPosition()` API が追加され、カーソル位置の制御が可能になった。ただし v2 は RC 段階

### 推奨対応
**現時点では保留**。理由:
- Bubble Tea v2 は RC であり、bubbles (textarea 等) のv2対応も不安定
- v1 での回避策が存在しない
- v2 が安定した段階で移行を検討

---

## 7. ステップ検索機能 [M] ✅ 実装済み

### 仕様
- `/` キーで検索モード（`ModeSearch`）に入る
- 画面下部に検索バーが表示される
- インクリメンタルに入力文字列にマッチする step をフィルタリング
- マッチ対象: step の ID + Title（大文字小文字区別なし）
- Enter で検索確定（現在のカーソル位置で Normal モードに戻る）
- Esc で検索キャンセル（元の表示に戻る）
- 検索中も j/k で候補間の移動が可能

### 変更内容

| ファイル | 変更内容 |
|----------|----------|
| `internal/tui/app.go` | `ModeSearch` 追加。`handleSearchMode` で textinput のイベント処理。`/` キーで検索モード開始 |
| `internal/tui/steplist.go` | `FilterByQuery(query string)` メソッド追加。Visible フラグを検索条件で更新。検索クリア用の `ClearFilter()` |
| `internal/tui/search.go` | 新規。`SearchBar` struct。`textinput.Model` をラップ。入力変更時にフィルタリングメッセージを発行 |
| `internal/tui/keymap.go` | `Search` キーバインド追加 (`/`) |
| `internal/tui/app.go` (View) | 検索モード時にステータスバーを検索バーに置き換え |

### 実装の詳細

```go
// ModeSearch: 検索入力を表示、stepList をリアルタイムフィルタ
// 検索バー: "/" + 入力文字列 + カーソル (画面最下部)
// フィルタロジック: strings.Contains(strings.ToLower(step.ID + " " + step.Title), query)
```

**フィルタの挙動:**
- マッチしない step は非表示（Visible = false）
- 親 step がマッチした場合、子も表示
- 子 step がマッチした場合、親も表示（パスを辿れるように）

---

## 実装順序

依存関係と難易度を考慮した推奨順序:

| 順番 | 項目 | 理由 |
|------|------|------|
| 1 | 項目5: ステップ幅拡張 | 独立・最小変更 |
| 2 | 項目1: delete/approve → viewed | 項目2 の前提。モデル変更を先に |
| 3 | 項目2: Conventional Comments | 項目1 のモデル変更後に実装 |
| 4 | 項目3: WezTerm サイズ動的化 | 独立・小規模 |
| 5 | 項目7: 検索機能 | 独立・中規模 |
| 6 | 項目4: Terminal.app spawner | 独立・中規模 |
| 7 | 項目6: IME 対応 | 保留 |

---

## スコープ外メモ

- `ActionInsertAfter`, `ActionInsertBefore` も項目1で削除対象（ユーザーの要件「基本的に modify のみ」に合致）
- テストの更新: 各項目で既存テストの修正が必要。golden file の更新含む
