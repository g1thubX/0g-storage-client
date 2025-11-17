#!/bin/bash

# 0G Storage 4GB文件上传测试脚本

echo "=== 0G Storage 4GB文件上传测试 ==="

# 检查必需的环境变量
check_env() {
    if [ -z "$1" ]; then
        echo "错误: 环境变量 $2 未设置"
        exit 1
    fi
}

check_env "$Testnet_RPC" "Testnet_RPC"
check_env "$Testnet_Indexer" "Testnet_Indexer" 
check_env "$PRIVATE_KEY" "PRIVATE_KEY"

echo "环境变量检查通过"
echo "RPC URL: $Testnet_RPC"
echo "Indexer URL: $Testnet_Indexer"
echo "Private Key: [已设置]"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装或不在PATH中"
    exit 1
fi

echo "Go 版本: $(go version)"

# 创建独立模块目录并编译
TEST_DIR="upload_test_$"
echo "创建测试目录: $TEST_DIR"

mkdir -p "$TEST_DIR"
cp large_file_upload_test.go "$TEST_DIR/"
cd "$TEST_DIR"

# 初始化Go模块
echo "初始化Go模块..."
go mod init upload_test

# 设置本地模块替换
go mod edit -replace github.com/0gfoundation/0g-storage-client=../

# 下载依赖
echo "下载依赖..."
go mod tidy

# 编译程序
echo "编译Go程序..."
if ! go build -o large_file_upload_test large_file_upload_test.go; then
    echo "编译失败"
    cd ..
    rm -rf "$TEST_DIR"
    exit 1
fi

echo "编译成功，开始运行测试..."

# 运行测试程序
./large_file_upload_test

# 清理
cd ..
echo "清理临时文件..."
rm -rf "$TEST_DIR"

echo "测试完成"