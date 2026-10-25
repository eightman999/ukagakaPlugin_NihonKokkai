# ukagakaPlugin_NihonKokkai の Ourin（macOS）移植

ambergon/ukagakaPlugin_NihonKokkai（Go の cgo c-shared 製 PLUGIN/2.0）を macOS 向けベースウェア Ourin で動かすための変更。
upstream のファイルの改変は、Windows 専用 API の分離とパス区切りだけに留めている。

## 変更点

- `code/main.go`: cgo の前置きから `#include <windows.h>` を外し、エクスポート関数の型を OS 非依存にした。
  - `load(h C.HGLOBAL, length C.long) C.BOOL` → `load(h unsafe.Pointer, length C.long) C.int`
  - `request(h C.HGLOBAL, length *C.long) C.HGLOBAL` → `request(h unsafe.Pointer, length *C.long) unsafe.Pointer`
  - `unload() bool` → `unload() C.int`（`int unload(void)` に合わせて 1 を返す。Windows でも BOOL 互換）
  - `C.GlobalFree` / `C.GlobalAlloc` の直接呼び出しを `freeHGlobal` / `allocHGlobal` に置き換えた。
- `code/hglobal_windows.go`（新規, `//go:build windows`）: 従来どおり `GlobalFree` / `GlobalAlloc(GPTR)` を使う。
- `code/hglobal_other.go`（新規, `//go:build !windows`）: `free` / `malloc` を使う（Ourin の ABI）。
- 設定ファイル等のパス組み立てを `filepath.Join` に変更した（`LoadJson.go` の `Config.json`）。
- `ourin/build.sh`, `ourin/build_nar.sh`, `ourin/README.md` を追加。`.gitignore` に `build/`, `main.dylib`, `libmain.dylib` を追加。

Ourin が読む ABI（Windows DLL と同じ形。HGLOBAL の代わりに malloc ポインタ、`long` は 64bit）:

- `int load(void *h, long len)`: `h` は malloc 確保の「プラグインのフォルダ/」（末尾スラッシュ付き）。モジュールが free する。
- `void *request(void *h, long *len)`: `h` は malloc 確保の要求文。モジュールが free する。応答は malloc で確保し、長さを `*len` に入れて返す（ホストが free）。
- `int unload(void)`

要求 1 行目は SSP と同じ `GET PLUGIN/2.0` / `NOTIFY PLUGIN/2.0`。

## ビルド

macOS（Xcode Command Line Tools と Go。Go 1.27.2 で確認）で:

```sh
ourin/build.sh        # arm64 / amd64 を CGO_ENABLED=1 で別々にビルドし、lipo で universal の main.dylib を直下に置く
ourin/build_nar.sh    # build/ukagakaPlugin_NihonKokkai-ourin.nar を作る（main.dylib を先に作っておく）
```

`build.sh` は環境変数 `GO` で go コマンドを差し替えられる（既定は PATH 上の `go`、無ければ `/opt/homebrew/bin/go`）。
中間生成物は `build/arm64/`, `build/amd64/`（`main.h` もここに出る）。

NAR には `descript.txt`, `install.txt`, `Config.json`（設定ひな形）, `main.dylib` だけを入れる。
`.git`, `build/`, `ourin/`, `code/`（Go ソース・bat・main.h）, `*.md`, Windows 版の `main.dll` は含めない。
`descript.txt` の `filename,main.dll` はそのまま。Ourin は同じフォルダの `main.dylib` を読む。

Windows 版は従来どおり `code/do.bat` でビルドできる（`hglobal_windows.go` が使われる）。
この移植作業では mingw が無いため Windows 向けのリンクまでは確認していない。代わりに、`windows.h` の型・関数だけを宣言したスタブヘッダと
`clang --target=i686-w64-windows-gnu` を CC にして `GOOS=windows GOARCH=386 CGO_ENABLED=1 go vet .` を通し、
Windows 側のファイル選択（`hglobal_windows.go` + `main.go`）と型検査が通ることを確認した。

## インストール

`build/ukagakaPlugin_NihonKokkai-ourin.nar` を Ourin にドロップするか、プラグインフォルダ（`descript.txt` のあるフォルダ）に `main.dylib` を置く。

## 設定

プラグインフォルダ直下の `Config.json`（upstream の ReadMe.md の説明どおり）。API キーは不要（国立国会図書館 国会会議録検索システム API）。

## 検証結果

ビルド: `ourin/build.sh` で universal（x86_64 + arm64）の `main.dylib` を作成（Go 1.27.2 / macOS 26 arm64）。

`nm -gU -arch arm64 main.dylib` / `-arch x86_64` とも `_load`, `_request`, `_unload` をエクスポートしている。
`codesign -v main.dylib` OK（ad-hoc 署名）。

`_tools/plugin_probe main.dylib <一時コピー> OnOurinNoSuchEvent`（Go 側の標準出力と C 側の出力は順序が前後する）:

```
load NihonKokkai
<plugin_dir>/
unload NihonKokkai
load(<plugin_dir>/) = 1
--- GET version
PLUGIN/2.0 204 No Content
Sender: NihonKokkai
Charset: UTF-8
--- GET OnGhostBoot
PLUGIN/2.0 204 No Content
Sender: NihonKokkai
Charset: UTF-8
--- GET OnSecondChange
PLUGIN/2.0 204 No Content
Sender: NihonKokkai
Charset: UTF-8
--- GET OnMenuExec
PLUGIN/2.0 204 No Content
Sender: NihonKokkai
Charset: UTF-8
--- GET OnOurinNoSuchEvent
PLUGIN/2.0 204 No Content
Sender: NihonKokkai
Charset: UTF-8
unload = 1
exit=0
```

`Config.json` を一時コピー側だけ `StartSec: 0`, `IntervalSec: 1`, `Words: "原発"`, `Human: "高市早苗"`, `From: -60`, `Until: -1` に変え、
scratchpad の `multi_probe -t 12`（1 秒ごとに OnSecondChange）で国会会議録検索 API（キー不要）への実通信と表示を確認した:

```
dlopen(main.dylib) = 0x77f74a40
load NihonKokkai
<plugin_dir>/
load(<plugin_dir>/) = 1
検索開始日2026-08-27
検索終了日2026-10-25
検索開始日2026-08-27
検索終了日2026-10-25
[main.dylib] OnSecondChange ->
PLUGIN/2.0 200 OK
Sender: NihonKokkai
Charset: UTF-8
Value: \0\b[2]\_q検索対象 : 高市早苗\n\_a[OnKokkaiUrl,https://kokkai.ndl.go.jp/txt/122205254X00120261005/23]2026-10-05 衆議院 本会議 第1号 #23 \_a\n\_q
unload NihonKokkai
unload(main.dylib) = 1
done
```

## 鍵・外部サービスが必要で未確認の機能

- API キーが必要な機能は無い。上記のとおり検索・`OnSecondChange` での表示は実通信で確認した。
- `OnKokkaiUrl`（リンククリック）は Ourin 上での実操作は未確認。Reference 無しで送ると upstream のコードが panic する。
- upstream の既知の問題（未修正）:
  - `Words` / `Human` に空要素（例: `"a,,b"` や空文字）があると `continue` で添字が進まず goroutine が無限ループする。
  - 通信エラー時に `resp` が nil のまま `resp.Body.Close()` するため panic し、ホストごと落ちる（`go vet` も `using resp before checking for errors` を報告）。

## Go 製プラグイン共通の注意（Ourin 側）

- **同一プロセスに複数の Go c-shared を読み込むことは可能だった**。ambergon 製 Go プラグイン 4 件（NewsAPI / NihonKokkai / simpleChatGPT / GoogleCalendar）を
  1 プロセスで dlopen → load → request（12 秒間 OnSecondChange を毎秒）しても問題なく、別途、GC・非同期プリエンプション（SIGURG）・goroutine を
  多数起こす Go c-shared を 3 個読み込み、8 本の C スレッドから同時に呼ぶ負荷試験も arm64（5 回）・x86_64（Rosetta, 1 回）とも完走した。
- **dlclose してはいけない**。macOS の dyld は Go の c-shared を実際にアンマップする（`_dyld_image_count` から消える）が、Go ランタイムのスレッドと
  シグナルハンドラは残るため、直後にシグナルハンドラがアンマップ済みアドレスへ飛び、プロセスが CPU 200% で固まった。
  Ourin の `Plugin`（PluginHost/Plugin.swift）は dlclose しないので現状は問題ない。プラグインの再読み込み・無効化を実装する場合も
  Go 製モジュールは閉じずに残すか、別プロセスで動かす必要がある（元の ReadMe にある「FreeLibrary でフリーズする」と同じ性質）。
- **Go の panic はホストごと落ちる**。エクスポート関数内やプラグインが起動した goroutine で panic すると、Go ランタイムがプロセスを abort する
  （例: NewsAPI に Reference 無しで `OnUrlSelectNewsAPI` を送ると `index out of range` で exit 134）。Ourin 本体を守るには、
  Go 製プラグインをプラグインごとに別プロセス（XPC サービス等）で動かし、要求文/応答文だけを受け渡すのが確実。

## ライセンス

upstream の ReadMe.md に `License: MIT` の記載あり（LICENSE ファイルは無い）。この移植部分も同じ MIT とする。
