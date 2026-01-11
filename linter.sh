#!/bin/bash

# MySQL N+1問題デモ - リンター実行スクリプト

set -e

echo "=== MySQL N+1問題デモ - 品質チェック ==="
echo ""

# Go モジュールの整理
echo "1. Go モジュールの整理..."
go mod tidy
echo "   ✓ go mod tidy 完了"

# フォーマットチェック
echo ""
echo "2. コードフォーマットチェック..."
if ! gofmt -l . | grep -v vendor | read; then
    echo "   ✓ フォーマット OK"
else
    echo "   ⚠ フォーマットが必要なファイル:"
    gofmt -l . | grep -v vendor
    echo ""
    echo "   gofmt -w . で修正してください"
fi

# go vet
echo ""
echo "3. go vet 実行..."
go vet ./...
echo "   ✓ go vet 完了"

# staticcheck (インストールされている場合)
echo ""
echo "4. staticcheck 実行..."
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
    echo "   ✓ staticcheck 完了"
else
    echo "   ⚠ staticcheck がインストールされていません"
    echo "   インストール: go install honnef.co/go/tools/cmd/staticcheck@latest"
fi

# golangci-lint (インストールされている場合)
echo ""
echo "5. golangci-lint 実行..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
    echo "   ✓ golangci-lint 完了"
else
    echo "   ⚠ golangci-lint がインストールされていません"
    echo "   インストール: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# ビルドテスト
echo ""
echo "6. ビルドテスト..."
go build -o /dev/null ./cmd/main.go
echo "   ✓ ビルド成功"

echo ""
echo "=== 品質チェック完了 ==="
