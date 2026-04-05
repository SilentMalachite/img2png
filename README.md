# img2png

[![Test](https://github.com/SilentMalachite/img2png/actions/workflows/test.yml/badge.svg)](https://github.com/SilentMalachite/img2png/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/SilentMalachite/img2png)](https://github.com/SilentMalachite/img2png/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8.svg?logo=go)](https://go.dev/)

JPEG・TIFF・WebP・BMP・GIF 画像を PNG に一括変換する CLI ツールです。  
*A CLI tool to batch-convert JPEG, TIFF, WebP, BMP, and GIF images to PNG.*

---

## 日本語

### 特徴

- TIFF・JPEG・WebP・BMP・GIF を **PNG に変換**
- **フォルダ**を渡すと再帰的にスキャンして **ZIP にまとめて出力**
- **単一ファイル**を渡すと同じフォルダに PNG を出力
- 同名ファイルは `photo.png` → `photo_2.png` と自動で連番付与
- ZIP 内はフラット構造（サブフォルダなし）
- **CGO 不使用**（macOS / Windows / Linux 対応の単一バイナリ）
- ファイルを **ドラッグ&ドロップ** するだけで使える

### ダウンロード

[Releases](../../releases) から対応するバイナリをダウンロードしてください。

| OS | ファイル |
|----|---------|
| macOS (Apple Silicon) | `img2png-darwin-arm64` |
| macOS (Intel) | `img2png-darwin-amd64` |
| Windows | `img2png-windows-amd64.exe` |
| Linux | `img2png-linux-amd64` |

**macOS の場合:** ダウンロード後、実行権限を付与してください。

```bash
chmod +x img2png-darwin-arm64
```

初回起動時に「開発元を確認できません」と表示される場合は、**システム設定 → プライバシーとセキュリティ** から許可してください。

### 使い方

#### ドラッグ&ドロップ

バイナリに画像ファイルまたはフォルダをドラッグ&ドロップするだけで変換が始まります。

#### コマンドライン

**フォルダを変換（ZIP 出力）:**

```bash
./img2png 写真フォルダ/
```

フォルダと同じ階層に `写真フォルダ.zip` が生成されます。

```
Done: 写真フォルダ.zip
```

ZIP 内の構成例:

```
写真フォルダ.zip
├── photo.png
├── photo_2.png   ← 別サブフォルダに同名ファイルがあった場合
└── scan.png
```

**単一ファイルを変換（PNG 直接出力）:**

```bash
./img2png photo.jpg
```

同じフォルダに `photo.png` が生成されます。

```
Done: photo.png
```

### 対応フォーマット

| 拡張子 | フォーマット |
|--------|------------|
| `.tif` `.tiff` | TIFF |
| `.jpg` `.jpeg` | JPEG |
| `.webp` | WebP |
| `.bmp` | BMP |
| `.gif` | GIF（アニメーションは最初のフレームのみ）|

拡張子の大文字小文字は問いません。

### エラー時の動作

エラーが発生した場合はメッセージを表示して一時停止します（ドラッグ&ドロップ利用時にウィンドウが閉じないよう）。

| 状況 | メッセージ |
|------|-----------|
| 対象ファイルが 0 件 | `no supported image files found` |
| 書き込み権限なし | `cannot write to directory: ...` |
| 変換失敗（個別） | `skipped: photo.jpg (unsupported or broken)` |

---

## English

### Features

- Converts **TIFF, JPEG, WebP, BMP, and GIF** images to PNG
- **Folder** input: recursively scans and outputs a **ZIP file**
- **Single file** input: outputs a PNG alongside the source file
- Duplicate filenames are automatically numbered: `photo.png`, `photo_2.png`, …
- ZIP contents are flat (no subdirectory structure)
- **No CGO** — single binary for macOS, Windows, and Linux
- Works with **drag-and-drop** — no terminal required

### Download

Download the binary for your platform from [Releases](../../releases).

| OS | File |
|----|------|
| macOS (Apple Silicon) | `img2png-darwin-arm64` |
| macOS (Intel) | `img2png-darwin-amd64` |
| Windows | `img2png-windows-amd64.exe` |
| Linux | `img2png-linux-amd64` |

**macOS:** After downloading, make the file executable:

```bash
chmod +x img2png-darwin-arm64
```

If macOS shows "cannot verify the developer", allow it in **System Settings → Privacy & Security**.

### Usage

#### Drag and Drop

Drag an image file or folder onto the binary to start conversion.

#### Command Line

**Convert a folder (ZIP output):**

```bash
./img2png photos/
```

Creates `photos.zip` in the same directory as the folder.

```
Done: photos.zip
```

Example ZIP contents:

```
photos.zip
├── photo.png
├── photo_2.png   ← duplicate basename from a subfolder
└── scan.png
```

**Convert a single file (PNG output):**

```bash
./img2png photo.jpg
```

Creates `photo.png` alongside the source file.

```
Done: photo.png
```

### Supported Formats

| Extension | Format |
|-----------|--------|
| `.tif` `.tiff` | TIFF |
| `.jpg` `.jpeg` | JPEG |
| `.webp` | WebP |
| `.bmp` | BMP |
| `.gif` | GIF (first frame only for animated GIFs) |

Extensions are matched case-insensitively.

### Error Handling

On error, a message is displayed and the program pauses before exiting (so the window stays open for drag-and-drop users).

| Situation | Message |
|-----------|---------|
| No supported files found | `no supported image files found` |
| No write permission | `cannot write to directory: ...` |
| Individual file failure | `skipped: photo.jpg (unsupported or broken)` |

### For Developers

#### Requirements

- Go 1.26 or later

#### Build

```bash
git clone https://github.com/SilentMalachite/img2png.git
cd img2png
go build -o img2png .
```

#### Test

```bash
go test -race ./...
```

#### Release

Push a tag starting with `v` to trigger GitHub Actions, which builds binaries for all platforms and uploads them to Releases automatically.

```bash
git tag v1.0.0
git push origin v1.0.0
```

---

## License

MIT — see [LICENSE](LICENSE).
