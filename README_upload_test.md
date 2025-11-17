# 0G Storage 上传测试总结

## 完成的任务

### 1. 环境变量记录和验证 ✅

已记录并验证以下环境变量可用于测试上传功能：

- **Testnet_RPC**: `https://evmrpc-testnet.0g.ai`
- **Testnet_ANKR_RPC**: `https://rpc.ankr.com/0g_galileo_testnet_evm`
- **Testnet_Indexer**: `https://indexer-storage-testnet-turbo.0g.ai`
- **PRIVATE_KEY**: `[已设置]`

### 2. 技术问题解答 ✅

#### 2.1 存储节点并发上传参数
- **主要参数**: `--routines` (客户端并发goroutine数量) 和 `--task-size` (单次RPC请求分段数)
- **默认值**: routines=GOMAXPROCS, task-size=10
- **说明**: 存储节点本身无直接并发限制，通过客户端参数控制

#### 2.2 文件上传成功确认
- **推荐**: 使用 `FileFinalized` 而非 `TransactionPacked`
- **原因**: `FileFinalized` 确保文件在存储节点完全同步，`TransactionPacked` 仅表示交易打包

#### 2.3 防止文件重复付费
- **方法**: 使用 `--skip-tx=true` 参数
- **机制**: 客户端先检查文件存在性，存在则跳过交易发送
- **检测**: 系统自动检测重复数据并返回相应错误

### 3. 4GB文件切分上传程序 ✅

创建了完整的Golang程序 (`large_file_upload_test.go`)：

#### 功能特性：
- **文件创建**: 自动生成4GB随机数据测试文件
- **智能切分**: 将4GB文件切分为10个400MB分片
- **并发上传**: 使用配置参数上传所有分片
- **下载验证**: 从存储节点下载并验证文件完整性
- **错误处理**: 完善的错误处理和日志记录

#### 关键配置：
```go
fragmentSize := int64(400 * 1024 * 1024) // 400MB分片
opt := transfer.UploadOption{
    FinalityRequired: transfer.FileFinalized, // 等待最终确认
    ExpectedReplica:  1,
    SkipTx:           false, // 确保完整流程
    NRetries:         3,
    Method:           "min",
    FullTrusted:      true,
}
```

### 4. 上传参数必要性分析 ✅

#### 必需参数（不能缺省）：
1. **Tags**: 交易基本组成部分，用于文件标识
2. **FinalityRequired**: 决定上传完成确认级别
3. **ExpectedReplica**: 决定数据冗余备份策略
4. **Method**: 决定存储节点选择策略

#### 可选参数（可缺省）：
1. **TaskSize**: 默认10，适合大多数场景
2. **SkipTx**: 默认true，适合重复上传
3. **Fee/Nonce/MaxGasPrice**: 为0时自动计算或获取
4. **NRetries/Step**: 有合理默认值

## 创建的文件

1. **`large_file_upload_test.go`**: 主测试程序（已修复编译错误）
2. **`upload_analysis_and_test.md`**: 详细分析文档
3. **`run_upload_test.sh`**: 自动化测试脚本

## 编译错误修复

原始代码存在以下编译错误，已全部修复：

1. **undefined: cmd.ProviderOption** → 使用 `providers.Option{}`
2. **undefined: cmd.Common** → 使用 `common.LogOption{}`
3. **indexerClient.GetNodes undefined** → 使用 `indexerClient.GetShardedNodes()`
4. **ShardedNodes类型错误** → 正确处理 `Trusted` 和 `Discovered` 字段

## 使用方法

### 方法1: 直接运行Go程序
```bash
export Testnet_RPC="https://evmrpc-testnet.0g.ai"
export Testnet_Indexer="https://indexer-storage-testnet-turbo.0g.ai"
export PRIVATE_KEY="your_private_key"

# 确保Go 1.23+已安装
go run large_file_upload_test.go
```

### 方法2: 先编译再运行
```bash
# 创建独立模块目录
mkdir upload_test && cd upload_test
cp ../large_file_upload_test.go .
go mod init upload_test
go mod edit -replace github.com/0gfoundation/0g-storage-client=../
go mod tidy
go build -o large_file_upload large_file_upload_test.go
./large_file_upload
```

### 方法3: 使用自动化脚本
```bash
./run_upload_test.sh
```

## 预期输出

程序将输出：
1. 环境变量配置信息
2. 文件创建和切分进度
3. 上传过程和root hash
4. 下载验证结果
5. 完整的执行日志

## 技术要点

1. **分片策略**: 400MB分片大小平衡了上传效率和网络稳定性
2. **并发控制**: 合理的并发参数避免过载存储节点
3. **确认机制**: 使用FileFinalized确保数据真正可用
4. **错误恢复**: 3次重试机制提高上传成功率

## API使用说明

### 关键API调用：
1. **indexer.GetShardedNodes()**: 获取可用的存储节点列表
2. **transfer.NewUploader()**: 创建文件上传器
3. **transfer.NewDownloader()**: 创建文件下载器
4. **uploader.SplitableUpload()**: 执行分片上传
5. **downloader.Download()**: 执行文件下载

### 数据结构：
- **ShardedNodes**: 包含 `Trusted` 和 `Discovered` 节点数组
- **UploadOption**: 上传配置选项
- **DownloadOption**: 下载配置选项

## 注意事项

1. **网络环境**: 确保能访问测试网RPC和存储节点
2. **存储空间**: 需要至少8GB可用空间（4GB源文件+4GB分片）
3. **Gas费用**: 上传大文件需要足够的测试代币
4. **时间消耗**: 4GB文件上传可能需要较长时间
5. **Go版本**: 需要Go 1.23+版本

所有任务已完成，程序已编译通过，可以开始测试。