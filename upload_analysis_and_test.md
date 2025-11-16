# 0G Storage 上传测试和参数说明

## 1. 环境变量记录

在完成上传测试任务时，使用了以下环境变量：

- **Testnet_RPC**: `https://evmrpc-testnet.0g.ai` - 区块链RPC节点地址
- **Testnet_ANKR_RPC**: `https://rpc.ankr.com/0g_galileo_testnet_evm` - 备用RPC节点地址
- **Testnet_Indexer**: `https://indexer-storage-testnet-turbo.0g.ai` - 存储节点索引器地址
- **PRIVATE_KEY**: `[已设置但隐藏]` - 测试钱包私钥

以上环境变量都是有效的，可以用来测试上传功能。

## 2. 技术问题解答

### 2.1 存储节点并发上传参数

存储节点关于调整接收上传请求并发数的参数主要通过以下方式控制：

1. **客户端并发控制**: 
   - `--routines` 参数：控制客户端并发上传的goroutine数量
   - 默认值：GOMAXPROCS（系统CPU核心数）

2. **任务大小控制**:
   - `--task-size` 参数：单次RPC请求上传的分段数量
   - 默认值：10

3. **分片大小控制**:
   - `--fragment-size` 参数：大文件分片大小
   - 默认值：4GB (4294967296字节)

存储节点本身没有直接的并发限制参数，主要通过客户端的`--routines`和`--task-size`来控制并发度。

### 2.2 文件上传成功确认

使用0g-storage-client时，确认文件上传成功的方式：

**推荐使用 `FileFinalized`**，理由如下：

- `TransactionPacked`: 仅表示交易被打包到区块中，但数据可能还未完全同步到存储节点
- `FileFinalized`: 表示文件已经在存储节点上完全同步并确认，是真正的上传完成

代码中的配置：
```go
finalityRequired := transfer.TransactionPacked
if uploadArgs.finalityRequired {
    finalityRequired = transfer.FileFinalized
}
```

### 2.3 防止文件重复付费

使用0g-storage-client时防止文件重复付费的方法：

1. **使用 `SkipTx` 参数**:
   ```bash
   --skip-tx=true
   ```
   - 如果文件已经存在于链上，跳过发送交易
   - 默认值：true

2. **检查文件存在性**:
   客户端会先检查文件是否已经上传：
   ```go
   info, err := checkLogExistence(ctx, uploader.clients, tree.Root())
   if !opt.SkipTx || info == nil {
       // 只有在文件不存在或SkipTx=false时才发送交易
   }
   ```

3. **重复数据检测**:
   系统会检测重复数据并返回相应错误：
   ```go
   var dataAlreadyExistsError = "Invalid params: root; data: already uploaded and finalized"
   ```

## 3. 上传参数必要性说明

在上传请求中，各参数的必要性分析：

### 必需参数（不能缺省）：

1. **Tags**:
   - **理由**: 虽然可以为空，但这是交易的基本组成部分，用于标识文件
   - **缺省影响**: 交易无法正确构造

2. **FinalityRequired**:
   - **理由**: 决定等待上传完成的确认级别，影响上传可靠性
   - **缺省影响**: 无法确定上传何时真正完成

3. **ExpectedReplica**:
   - **理由**: 决定文件需要复制到多少个节点，影响数据可靠性
   - **缺省影响**: 无法保证数据的冗余备份

4. **Method**:
   - **理由**: 决定选择存储节点的策略，影响上传性能和成功率
   - **缺省影响**: 无法选择合适的存储节点

### 可选参数（可以缺省或有合理默认值）：

1. **TaskSize**:
   - **可以缺省**: 有默认值10，适合大多数场景

2. **SkipTx**:
   - **可以缺省**: 有默认值true，适合重复上传场景

3. **Fee**:
   - **可以缺省**: 为0时使用链上计算的费用

4. **Nonce**:
   - **可以缺省**: 为0时自动从链上获取

5. **MaxGasPrice**:
   - **可以缺省**: 为0时不限制gas价格

6. **NRetries**:
   - **可以缺省**: 有默认值0，不重试

7. **Step**:
   - **可以缺省**: 有默认值15，适用于gas价格调整

## 4. 4GB文件切分上传程序

已创建 `large_file_upload_test.go` 程序，功能包括：

1. **创建4GB测试文件**: 使用随机数据填充
2. **文件切分**: 将4GB文件切分为10个400MB的分片
3. **并发上传**: 使用配置的参数上传所有分片
4. **下载验证**: 下载上传的文件并验证完整性

### 关键配置：

```go
// 分片大小设置
fragmentSize := int64(400 * 1024 * 1024) // 400MB

// 上传选项
opt := transfer.UploadOption{
    Tags:             hexutil.MustDecode("0x"),
    FinalityRequired: transfer.FileFinalized, // 等待最终确认
    TaskSize:         10,
    ExpectedReplica:  1,
    SkipTx:           false, // 确保完整上传流程
    NRetries:         3,
    Step:             15,
    Method:           "min",
    FullTrusted:      true,
}
```

### 运行方式：

```bash
# 确保环境变量已设置
export Testnet_RPC="https://evmrpc-testnet.0g.ai"
export Testnet_Indexer="https://indexer-storage-testnet-turbo.0g.ai"
export PRIVATE_KEY="your_private_key"

# 运行程序
go run large_file_upload_test.go
```

程序会自动记录使用的环境变量，执行完整的文件切分、上传、下载流程，并输出详细的执行日志。