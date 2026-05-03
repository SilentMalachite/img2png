# img2png GUI 機能 設計書 (v1)

- **作成日**: 2026-05-03
- **対象**: img2png に Windows / macOS 対応の GUI 機能を追加する
- **ステータス**: ブレインストーミング合意済み、実装計画策定前

## 1. 目的とスコープ

img2png は現在 Go 製の CLI ツールで、JPEG/TIFF/WebP/BMP/GIF を PNG に一括変換する。本仕様では、既存の CLI 体験を壊さずに GUI モードを追加する。

### v1 で実装する機能

GUI ウィンドウから次が可能になる。

- ウィンドウへのファイル / フォルダのドラッグ&ドロップ受付
- 複数ファイル / フォルダの一括追加 (リスト形式)
- 出力先フォルダの指定 (空欄なら現行 CLI と同じ既定挙動)
- フォルダ入力時の出力モード切替: ZIP / 個別 PNG
- 同名ファイルの上書きポリシー: 連番付与 / 上書き / スキップ
- 変換中のキャンセル
- 言語切替 (日本語 / 英語、起動時に OS ロケール自動検出)
- 設定の永続化 (出力先 / 出力モード / ポリシー / 言語 / ウィンドウサイズ)

### v1 スコープ外

| 項目 | 理由 / 残置メモ |
|---|---|
| コード署名・公証 | macOS は Apple Developer Program (年99ドル)、Windows は EV 証明書必要。README 既記載の Security & Privacy 回避手順を継続 |
| インストーラ (DMG / MSI) | ZIP 配布で十分とする |
| リサイズ・圧縮レベル・カラープロファイル | ブレストで明示的に除外 |
| 並列変換 | 順次処理で十分、ブレストで明示的に除外 |
| 変換履歴 | ブレストで明示的に除外 |
| 更新確認・自動アップデート | スコープ外 |
| macOS の "Open With" 関連付け | 手動 D&D とファイル選択で十分 |
| Linux GUI の AppImage / Flatpak / Snap | tar.xz 配布のみ |
| CLI モードの GUI 統合 | CLI コードは現状維持、`internal/job` への移行は v2 以降 |
| GIF アニメーション全フレーム対応 | 既存 CLI と同じく最初のフレームのみ |

### 将来作業

- v2 候補: コード署名・公証、リサイズ機能、並列変換、CLI モードを `internal/job` に統合
- v3 候補: インストーラ、自動アップデート、ファイル関連付け

## 2. 主要な技術判断

| 決定事項 | 内容 |
|---|---|
| GUI フレームワーク | **Fyne v2** (`fyne.io/fyne/v2`) |
| CGO | **CGO=1 必須に変更**。README の「CGO 不使用」記述は撤回 |
| バイナリ構成 | **同一バイナリ**。引数なし=GUI、引数あり=既存 CLI |
| ウィンドウレイアウト | **左右分割**: 左ペイン=ファイルリスト、右ペイン=設定パネル |
| 配布パッケージ | **ネイティブパッケージ**: macOS は `.app` バンドル、Windows はアイコン埋込み `.exe`、ZIP 配布 |

### 仮定とリスク

- **仮定1**: バイナリサイズ約 30MB (現状数 MB から増加) はユーザーに受容される
- **仮定2**: macOS 未署名アプリの初回起動は既存 README の手順で運用可能
- **仮定3**: Fyne v2 の API 安定性は v1 リリース期間中維持される
- **リスク1**: macOS Sequoia 以降で未署名アプリ起動制限がさらに厳しくなる可能性 → 顕在化時は v2 でコード署名対応へ昇格
- **リスク2**: Fyne のドラッグ&ドロップ API がプラットフォーム間で挙動差異あり → 手動検証で確認、必要なら + ボタン (ファイルダイアログ) でフォールバック

## 3. アーキテクチャ概要

### バイナリ起動時のディスパッチ

```
img2png 起動
   │
   ├─ 引数あり → 既存 CLI フロー (runFile / runDir)
   │              ※ 現状の挙動を 100% 維持
   │
   └─ 引数なし → GUI モード (新規)
                  Fyne でメインウィンドウを開く
                  内部的には converter / archiver を呼び出す
```

ターミナルから引数なしで `img2png` を叩いた場合は GUI が起動する。CLI の使い方を見る手段として `img2png --help` を新規追加する (現行の `Usage:` メッセージは廃止)。

### パッケージ構成

```
img2png/
├── main.go                    (ディスパッチ + 既存 CLI ロジック)
├── internal/
│   ├── converter/             (変更なし、既存を GUI から流用)
│   ├── archiver/              (変更なし、既存を GUI から流用)
│   ├── gui/                   ★ 新規: Fyne アプリ本体
│   │   ├── app.go             (アプリ初期化、メインウィンドウ組み立て)
│   │   ├── filelist.go        (左ペイン: D&D、ファイルリスト管理)
│   │   ├── settings_panel.go  (右ペイン: 出力先・モード・ポリシー)
│   │   ├── progress.go        (進捗バー、キャンセル)
│   │   └── menu.go            (メニュー: 言語切替、About 等)
│   ├── job/                   ★ 新規: 変換ジョブのオーケストレーション
│   │   ├── job.go             (ジョブ定義、進捗イベント)
│   │   └── runner.go          (goroutine で変換実行、context.Cancel でキャンセル)
│   ├── settings/              ★ 新規: 設定の永続化
│   │   └── settings.go        (JSON 読み書き、OS 標準パス)
│   └── i18n/                  ★ 新規: ラベル翻訳
│       └── i18n.go            (日本語 / 英語、OS ロケール自動検出)
├── assets/
│   └── icon.png               ★ 新規: 512x512 アプリアイコン
└── FyneApp.toml               ★ 新規: アプリ名 / バージョン / Bundle ID
```

### 設計上の重要ポイント

- `converter` / `archiver` パッケージは**変更しない** (既存 CLI の動作を保証)
- GUI は `internal/gui` に隔離し、CLI モードからは触らない
- 変換実行は `internal/job` に集約。v1 では GUI のみが利用、CLI は現状維持
- バイナリサイズは GUI 追加で約 30MB 前後に増加

## 4. コンポーネント詳細

### 4.1 メインウィンドウ (`internal/gui/app.go`)

`container.NewHSplit` で左右分割。最小サイズ 640x400、起動時に前回サイズを復元。

### 4.2 左ペイン: ファイルリスト (`internal/gui/filelist.go`)

```go
type FileList struct {
    items    []FileItem  // 追加されたファイル / フォルダ
    onChange func()      // 変更通知
}

type FileItem struct {
    Path   string
    IsDir  bool
    Status Status  // Pending / Running / Done / Failed / Skipped
}
```

- 上部: D&D 受付エリア (`fyne.Window.SetOnDropped` でハンドル)
- 中央: `widget.List` でファイル一覧。各行に状態アイコン、ファイル名、削除ボタン
- 下部ボタン: 「+ ファイル追加」「+ フォルダ追加」「クリア」(いずれもファイルダイアログ経由)
- 同一パスの重複追加は無視

### 4.3 右ペイン: 設定パネル (`internal/gui/settings_panel.go`)

- **出力先フォルダ** — `widget.Entry` + 「参照」ボタン (フォルダダイアログ)。空欄なら**現行 CLI と同じ既定**: ファイル入力時は同フォルダ、フォルダ入力時は親フォルダ
- **フォルダ入力時の出力モード** — `widget.RadioGroup`: ZIP / 個別 PNG (既定: ZIP)
- **同名ファイル時のポリシー** — `widget.Select`: 連番付与 / 上書き / スキップ (既定: 連番付与 = 現状動作)

設定変更は即座に内部状態へ反映。永続化は次回起動時の復元用。

### 4.4 進捗エリア (`internal/gui/progress.go`)

ファイルリストの直下、変換中のみ表示する。

- 全体進捗バー (n / total ファイル完了)
- 現在処理中のファイル名表示
- 「キャンセル」ボタン
- 完了後は結果サマリ表示 (成功 N / スキップ M / 失敗 K) と「出力フォルダを開く」ボタン

### 4.5 メニュー (`internal/gui/menu.go`)

- macOS: ネイティブメニューバー (Fyne が自動で Dock アイコンと統合)
- Windows: Fyne 標準のウィンドウメニューバー
- 項目: 言語切替 (日本語 / 英語 / システムに従う)、About、終了

### 4.6 ジョブランナー (`internal/job/runner.go`)

```go
type Job struct {
    Items      []FileItem
    OutputDir  string           // 空なら入力に応じて自動決定
    OutputMode Mode             // Zip / Individual
    Overwrite  OverwritePolicy  // Increment / Overwrite / Skip
}

func Run(ctx context.Context, j Job, progress chan<- Event) error
```

- 1 ファイルずつ順次変換 (`converter.ConvertFile` を再利用)
- 各ファイル完了 / 失敗で `progress` チャネルに `Event` 送信
- `ctx.Done()` でキャンセル: 進行中のファイルは完了させ、新規ファイルの処理を停止。完了済みファイルはそのまま残す
- フォルダ入力 + ZIP モードの場合: 一時ディレクトリに各 PNG 出力 → 全完了後に ZIP 化 (既存 `runDir` と同じ流れ)。**キャンセル時は少なくとも 1 ファイル成功していれば ZIP を生成、0 件ならエラー扱い**

### 4.7 設定の永続化 (`internal/settings/settings.go`)

保存先 (OS 標準):

- macOS: `~/Library/Application Support/img2png/settings.json`
- Windows: `%APPDATA%\img2png\settings.json`

保存項目: 出力先、出力モード、上書きポリシー、言語、ウィンドウサイズ。`os.UserConfigDir()` で取得。読み込み失敗時は既定値で起動 (壊れたファイルは黙って無視 + ログ)。

### 4.8 国際化 (`internal/i18n/i18n.go`)

- 単純な `map[string]string` ベース。キー → ロケール別ラベル
- 起動時の優先順: 設定ファイルの言語 > OS ロケール自動検出 > 英語フォールバック
- v1 は日本語・英語のみ。`golang.org/x/text` 等は使わずミニマムで実装

## 5. データフロー / 主要シナリオ

### シナリオ A: 起動からファイル追加まで

```
ユーザー img2png.app をダブルクリック (または img2png.exe 実行)
  ↓
main(): len(os.Args) == 1 → gui.Run() 呼び出し
  ↓
gui.Run(): Fyne App + 設定読み込み + i18n 初期化 + メインウィンドウ表示
  ↓
ユーザー: ファイルをウィンドウへドロップ / + ボタンでダイアログ選択
  ↓
FileList: パス検証 (存在チェック・拡張子チェック・重複除去)
  ↓
リスト更新 + 「変換開始」ボタン活性化
```

### シナリオ B: 変換実行 (正常系)

```
「変換開始」クリック
  ↓
gui: 設定パネルの値を読み取り Job 構造体を構築
  ↓
gui: ボタン無効化、進捗エリア表示、ctx, cancel := context.WithCancel(...)
  ↓
job.Run(ctx, job, progressCh) を別 goroutine で起動
  ↓
job: 各 FileItem について
       - Dir なら walk → 各画像を converter.ConvertFile
         (ZIP モードなら一時ディレクトリへ、個別 PNG モードなら出力先へ直接)
       - File なら converter.ConvertFile
       - 1 件ごとに Event{Path, Status, Error} を progressCh へ送信
  ↓
gui: progressCh を受信し UI を更新 (Fyne の fyne.Do() でメインスレッドへ)
  ↓
全完了 → ZIP モードなら archiver.Archive で zip 化 → close(progressCh)
  ↓
gui: 結果サマリ表示、「出力フォルダを開く」ボタン表示、ボタン群を再活性化
```

### シナリオ C: キャンセル

```
変換中に「キャンセル」クリック
  ↓
gui: cancel() を呼び出し
  ↓
job: 現在処理中のファイル変換は完了させる (converter 呼び出しはアトミック)
     次のファイルに進む前に ctx.Err() を確認 → 中断
  ↓
ZIP モード: 完了済み PNG が ≥ 1 なら ZIP 生成、0 件ならエラー
個別 PNG モード: 完了済みファイルはそのまま残置
  ↓
gui: 「キャンセルされました (成功 N / 未処理 M)」表示
```

### シナリオ D: エラー処理

| 状況 | 挙動 |
|---|---|
| 出力先に書き込み権限なし | 変換開始時に検出 → モーダルダイアログでエラー表示、変換は開始しない |
| 個別ファイル変換失敗 (壊れた画像など) | リスト上で「スキップ」マーク、変換は続行、サマリにスキップ件数を含める |
| 全ファイル失敗 | 完了時にエラーメッセージダイアログ |
| 出力先に既存 PNG があり「スキップ」ポリシー | リスト上で「スキップ」マーク、続行 |
| 出力先に既存 PNG があり「上書き」ポリシー | 黙って上書き |
| ZIP 化失敗 | エラーダイアログ、ただし一時ディレクトリの中間 PNG はクリーンアップ |

### シナリオ E: 既存 CLI ユーザーへの非影響保証

- `img2png photo.jpg` → 引数あり → 既存 `run()` 関数へ。出力・終了挙動は完全に現行通り
- `img2png` をターミナルから引数なしで叩いた場合 → GUI モード起動
- `img2png --help` (新規追加) で CLI の使い方を表示

### シナリオ F: D&D (ウィンドウ + バイナリ両対応)

- **バイナリへの D&D** (現状機能・OS 仕様): OS が `os.Args` に渡す → 引数あり → CLI モードで動作 (現状通り)
- **GUI ウィンドウへの D&D** (新規): `fyne.Window.SetOnDropped` でハンドル → ファイルリストへ追加

両方の方法をユーザーが選べるようにする。

## 6. テスト戦略

### 既存テストの保証 (変更なし)

- `internal/converter/converter_test.go` と `internal/archiver/archiver_test.go` はそのまま維持
- CI で現行通り `go test -race ./...` がパスすること

### 新規パッケージのユニットテスト

| パッケージ | テスト対象 | 方針 |
|---|---|---|
| `internal/job` | `Run()` の進捗イベント、キャンセル時の挙動、ZIP モード / 個別 PNG モード分岐、上書きポリシー | 一時ディレクトリと固定の小さな画像で実行。`context.WithCancel` で中断、イベント受信を検証 |
| `internal/settings` | JSON 保存 / 読込、破損ファイルの扱い、未存在時の既定値 | `t.TempDir()` で隔離、`os.UserConfigDir` はテスト用に注入可能に |
| `internal/i18n` | キー解決、ロケール検出、フォールバック | テーブル駆動 |

### GUI ロジックのテスト方針

Fyne のウィジェットを直接テストするのはコストが高いので、**ロジックを純粋関数として切り出してテスト**する。

- ファイルリストの追加 / 削除 / 重複除去 → `filelist.go` の純粋関数を直接テスト
- 設定パネル → 設定パネルから `Job` 構造体を組み立てる純粋関数 `BuildJob(panelState) Job` をテスト
- ウィジェットそのもの (描画、イベントハンドラ) はテスト対象外 (手動検証)

### 手動検証チェックリスト

- [ ] macOS で `.app` をダブルクリック起動 → ウィンドウ表示
- [ ] Windows で EXE をダブルクリック起動 → ウィンドウ表示
- [ ] ウィンドウへのファイル D&D (単一・複数・フォルダ)
- [ ] 変換実行 → 進捗表示 → 完了サマリ
- [ ] キャンセル → 中断、完了済みファイルが残る
- [ ] 既存 CLI モード `img2png photo.jpg` の動作
- [ ] 設定の永続化 (再起動後に復元)
- [ ] 言語切替 (日 / 英)
- [ ] エラー: 壊れた画像、書き込み権限なし

### CI でのテスト実行

- 現行の matrix (ubuntu / macos / windows) で `go test -race ./...` を維持
- **CGO=1 のテスト**: macOS と Windows の runner はネイティブコンパイラが入っているのでそのまま動く。Linux は `libgl1-mesa-dev` 等の依存追加が必要 → CI にステップ追加
- GUI 起動を伴うテストは行わない (ヘッドレス実行が複雑なため、ロジックのテストで担保)

## 7. 配布 / CI 変更

### 現状の課題

現在の `.github/workflows/build.yml` は `ubuntu-latest` 1 台で `CGO_ENABLED=0` を使い 4 プラットフォーム分をクロスコンパイルしている。Fyne は CGO 必須なのでこの方針は使えない。

### 変更後のビルド戦略

各プラットフォームのネイティブ runner でビルドする。

```yaml
strategy:
  matrix:
    include:
      - runner: macos-14          # Apple Silicon
        goos: darwin
        goarch: arm64
        artifact: img2png-darwin-arm64
        package: app

      - runner: macos-13          # Intel
        goos: darwin
        goarch: amd64
        artifact: img2png-darwin-amd64
        package: app

      - runner: windows-latest
        goos: windows
        goarch: amd64
        artifact: img2png-windows-amd64.exe
        package: exe

      - runner: ubuntu-latest
        goos: linux
        goarch: amd64
        artifact: img2png-linux-amd64
        package: tar.xz
        deps: libgl1-mesa-dev xorg-dev
```

### 主要なステップ追加

```yaml
- name: Install Linux deps
  if: matrix.goos == 'linux'
  run: sudo apt-get update && sudo apt-get install -y ${{ matrix.deps }}

- name: Install fyne CLI
  run: go install fyne.io/fyne/v2/cmd/fyne@latest

- name: Build with packaging
  env:
    CGO_ENABLED: '1'
  run: fyne package -os ${{ matrix.goos }} -name img2png -icon assets/icon.png

- name: Zip artifact (macOS .app)
  if: matrix.goos == 'darwin'
  run: zip -r ${{ matrix.artifact }}.zip img2png.app
```

### 配布物 (リリースに添付)

| OS | アーティファクト | 内容 |
|---|---|---|
| macOS Apple Silicon | `img2png-darwin-arm64.zip` | `img2png.app` バンドル (アイコン・Info.plist 込み) |
| macOS Intel | `img2png-darwin-amd64.zip` | 同上 |
| Windows | `img2png-windows-amd64.zip` | `img2png.exe` (アイコン埋込み) |
| Linux | `img2png-linux-amd64.tar.xz` | `img2png` バイナリ (依存ライブラリは別途要) |

### Info.plist / アイコンの管理

- `assets/icon.png` (512x512、新規追加) — Fyne が macOS の `.icns`、Windows の `.ico` を自動生成
- `FyneApp.toml` (新規追加) — アプリ名、バージョン、Bundle ID (例: `com.silentmalachite.img2png`) を定義
- バージョンは git tag から取得 (既存の `v*` タグトリガを継続)

### README 更新

- 「CGO 不使用」の記述を**削除** (あるいは「CLI モードのみ CGO 不使用相当」に書き換え)
- macOS 初回起動の許可手順は既存のまま (むしろ GUI で重要度が増す)
- 配布ファイル名を `.zip` / `.tar.xz` に更新
- GUI 使用方法の節を追加 (D&D、ボタン、設定)

### Linux の扱い

- Linux は GUI ビルドを CI で生成はするが、ランタイム依存 (`libgl1`、`libx11` 等) の事前インストールがエンドユーザに必要となるため、README で注意喚起
- 既存の CLI Linux ユーザは `img2png photo.jpg` のように引数渡しで使い続けられる
