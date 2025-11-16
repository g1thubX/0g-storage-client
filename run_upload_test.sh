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

# 编译Go程序
echo "编译Go程序..."
if ! go build -o large_file_upload_test large_file_upload_test.go; then
    echo "编译失败"
    exit 1
fi

echo "编译成功，开始运行测试..."

# 运行测试程序
./large_file_upload_test

echo "测试完成"