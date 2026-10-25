#!/bin/sh
# Ourin 用の NAR（zip）を作る。リポジトリ管理用のファイル、Go ソース、Windows 版 DLL は含めない。
# 先に ourin/build.sh で main.dylib を作っておくこと。
set -eu
cd "$(dirname "$0")/.."
[ -f main.dylib ] || { echo "main.dylib がありません。先に ourin/build.sh を実行してください。" >&2; exit 1; }
NAME=$(basename "$(pwd)")
mkdir -p build
rm -f "build/$NAME-ourin.nar"
zip -qr "build/$NAME-ourin.nar" . \
	-x '.git/*' 'build/*' 'ourin/*' 'code/*' '.github/*' '*.md' '*.go' '.gitignore' '.DS_Store' 'main.dll'
unzip -l "build/$NAME-ourin.nar"
echo "build/$NAME-ourin.nar"
