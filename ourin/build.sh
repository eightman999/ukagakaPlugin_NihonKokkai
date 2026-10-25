#!/bin/sh
# macOS（Ourin）向けに Go の c-shared ライブラリを arm64 / amd64 で別々にビルドし、
# lipo で universal の main.dylib にまとめてプラグインフォルダ直下へ置く。
# descript.txt の filename は main.dll のまま。Ourin は同じフォルダの main.dylib を読む。
set -eu
cd "$(dirname "$0")/.."
ROOT=$(pwd)
GO=${GO:-$(command -v go || echo /opt/homebrew/bin/go)}
OUT="$ROOT/build"

for arch in arm64 amd64; do
	echo "build darwin/$arch"
	mkdir -p "$OUT/$arch"
	(cd "$ROOT/code" && CGO_ENABLED=1 GOOS=darwin GOARCH=$arch \
		"$GO" build -trimpath -buildmode=c-shared -o "$OUT/$arch/main.dylib" .)
done

lipo -create -output "$ROOT/main.dylib" "$OUT/arm64/main.dylib" "$OUT/amd64/main.dylib"
# 各スライスはリンカが ad-hoc 署名済み。universal 全体もまとめて ad-hoc 署名しておく。
codesign --force --sign - "$ROOT/main.dylib"
lipo -info "$ROOT/main.dylib"
