# 0G Storage 测试网上传文件环境变量配置指南

## 1. 使用测试链上传文件需要设置的环境变量参数

### 必需参数

使用 0G Storage 测试网上传文件时，以下参数是**必需的**：

#### 1.1 区块链连接参数
- **`--url`**: 区块链 RPC 节点地址
  - 用途：与 0G Storage 智能合约交互
  - 格式：完整的 HTTP/HTTPS URL
  - 示例：`https://testnet-rpc.0g.ai` 或 `http://127.0.0.1:8545`

- **`--key`**: 私钥
  - 用途：签名交易并与智能合约交互
  - 格式：64位十六进制字符串（不带 0x 前缀）
  - 示例：`46b9e861b63d3509c88b7817275a30d22d62c8cd8fa6486ddee35ef0d8e0495f`
  - ⚠️ **安全提示**: 请勿在生产环境中直接使用命令行传递私钥

#### 1.2 文件参数
- **`--file`**: 要上传的文件路径
  - 用途：指定本地文件位置
  - 格式：相对或绝对路径
  - 示例：`./test.txt` 或 `/home/user/data.bin`

#### 1.3 存储节点参数（二选一）

**方式一：通过 Indexer（推荐）**
- **`--indexer`**: 存储节点索引器地址
  - 用途：自动发现和选择最优存储节点
  - 格式：HTTP/HTTPS URL
  - 示例：`https://testnet-indexer.0g.ai`

**方式二：直接指定节点**
- **`--node`**: 存储节点 RPC 地址列表
  - 用途：直接指定要使用的存储节点
  - 格式：逗号分隔的 URL 列表
  - 示例：`http://node1.0g.ai:5678,http://node2.0g.ai:5678`

### 可选参数

#### 2.1 交易相关参数
- **`--fee`**: 交易费用（单位：a0gi，1 a0gi = 10^-18 0G）
  - 默认值：0（使用链上计算的费用）
  - 示例：`--fee 0.001`

- **`--nonce`**: 交易 nonce
  - 默认值：0（自动从链上获取）
  - 示例：`--nonce 10`

- **`--gas-price`**: 自定义 Gas 价格
  - 全局参数，适用于所有交易
  - 示例：`--gas-price 10000000000`

- **`--gas-limit`**: 自定义 Gas 限制
  - 全局参数，适用于所有交易
  - 示例：`--gas-limit 1000000`

- **`--max-gas-price`**: 最大 Gas 价格限制
  - 防止 Gas 价格过高
  - 示例：`--max-gas-price 50000000000`

#### 2.2 上传行为参数
- **`--tags`**: 文件标签
  - 默认值：`0x`
  - 格式：十六进制字符串
  - 示例：`--tags 0x1234567890abcdef`

- **`--expected-replica`**: 期望的副本数量
  - 默认值：1
  - 示例：`--expected-replica 3`

- **`--skip-tx`**: 如果交易已存在，跳过发送
  - 默认值：true
  - 示例：`--skip-tx=false`

- **`--finality-required`**: 等待文件在节点上达到最终确认状态
  - 默认值：false
  - 示例：`--finality-required=true`

- **`--task-size`**: 单次 RPC 请求上传的分段数量
  - 默认值：10
  - 示例：`--task-size 20`

#### 2.3 性能和重试参数
- **`--routines`**: 并发上传的 goroutine 数量
  - 默认值：GOMAXPROCS（系统 CPU 核心数）
  - 示例：`--routines 8`

- **`--fragment-size`**: 大文件分片大小（字节）
  - 默认值：4294967296 (4GB)
  - 示例：`--fragment-size 1073741824`

- **`--n-retries`**: 上传失败时的重试次数
  - 默认值：0
  - 示例：`--n-retries 3`

- **`--step`**: Gas 价格增长步长（step/10）
  - 默认值：15（每次增长 1.5 倍）
  - 示例：`--step 20`

- **`--timeout`**: CLI 任务超时时间
  - 默认值：0（无超时）
  - 格式：Go duration 格式
  - 示例：`--timeout 5m`

#### 2.4 节点选择参数
- **`--method`**: 节点选择方法
  - 可选值：`min`, `max`, `random`, 或正数
  - 默认值：`min`
  - 示例：`--method random`

- **`--full-trusted`**: 是否仅使用完全信任的节点
  - 默认值：true
  - 示例：`--full-trusted=false`

#### 2.5 全局参数
- **`--log-level`**: 日志级别
  - 默认值：`info`
  - 可选值：`debug`, `info`, `warn`, `error`
  - 示例：`--log-level debug`

- **`--log-color-disabled`**: 禁用彩色日志
  - 默认值：false
  - 示例：`--log-color-disabled`

- **`--web3-log-enabled`**: 启用 Web3 RPC 日志
  - 默认值：false
  - 示例：`--web3-log-enabled`

- **`--rpc-timeout`**: 单次 RPC 请求超时
  - 默认值：30s
  - 示例：`--rpc-timeout 60s`

- **`--rpc-retry-count`**: RPC 请求重试次数
  - 默认值：5
  - 示例：`--rpc-retry-count 10`

- **`--rpc-retry-interval`**: RPC 请求重试间隔
  - 默认值：5s
  - 示例：`--rpc-retry-interval 10s`

### 完整示例

#### 示例 1：使用 Indexer 上传文件（推荐）
```bash
./0g-storage-client upload \
  --url https://testnet-rpc.0g.ai \
  --key YOUR_PRIVATE_KEY \
  --indexer https://testnet-indexer.0g.ai \
  --file ./myfile.txt \
  --skip-tx=false
```

#### 示例 2：直接指定存储节点上传
```bash
./0g-storage-client upload \
  --url https://testnet-rpc.0g.ai \
  --key YOUR_PRIVATE_KEY \
  --node http://node1.0g.ai:5678,http://node2.0g.ai:5678 \
  --file ./myfile.txt \
  --expected-replica 2
```

#### 示例 3：高级配置上传
```bash
./0g-storage-client upload \
  --url https://testnet-rpc.0g.ai \
  --key YOUR_PRIVATE_KEY \
  --indexer https://testnet-indexer.0g.ai \
  --file ./largefile.bin \
  --expected-replica 3 \
  --finality-required=true \
  --routines 16 \
  --n-retries 5 \
  --max-gas-price 100000000000 \
  --log-level debug \
  --timeout 10m
```

### KV 写入参数（kv-write）

如果需要使用 KV 存储功能，需要以下参数：

#### 必需参数
- **`--url`**: 区块链 RPC 地址
- **`--key`**: 私钥
- **`--stream-id`**: 流 ID（格式：0x...）
- **`--stream-keys`**: KV 键列表（逗号分隔）
- **`--stream-values`**: KV 值列表（逗号分隔）

#### 可选参数
- **`--indexer`** 或 **`--node`**: 存储节点地址
- **`--version`**: 键版本（默认：MaxUint64）
- **`--expected-replica`**: 期望副本数（默认：1）
- **`--skip-tx`**: 跳过已存在的交易（默认：false）
- **`--method`**: 节点选择方法（默认：random）

#### KV 写入示例
```bash
./0g-storage-client kv-write \
  --url https://testnet-rpc.0g.ai \
  --key YOUR_PRIVATE_KEY \
  --indexer https://testnet-indexer.0g.ai \
  --stream-id 0x123456... \
  --stream-keys key1,key2,key3 \
  --stream-values value1,value2,value3
```

---

## 2. 如何获取完成后的代码

### 当前分支信息

我已经在预先创建好的分支上工作：

**分支名称**: `feat/testnet-upload-env-and-code-delivery-perms`

### 检查分支状态

您可以使用以下命令检查分支：

```bash
# 查看所有分支
git branch -a

# 查看当前分支和状态
git status

# 查看分支的提交历史
git log --oneline

# 查看与远程分支的差异
git diff main...feat/testnet-upload-env-and-code-delivery-perms
```

### 获取完成的代码

#### 方法 1：拉取分支（推荐）
如果分支已经推送到远程仓库：
```bash
# 切换到该分支
git checkout feat/testnet-upload-env-and-code-delivery-perms

# 拉取最新更改
git pull origin feat/testnet-upload-env-and-code-delivery-perms
```

#### 方法 2：查看更改内容
查看该分支相对于 main 分支的所有更改：
```bash
# 查看文件差异
git diff main..feat/testnet-upload-env-and-code-delivery-perms

# 查看修改的文件列表
git diff --name-only main..feat/testnet-upload-env-and-code-delivery-perms

# 查看详细的提交记录
git log main..feat/testnet-upload-env-and-code-delivery-perms --oneline
```

#### 方法 3：创建补丁文件
如果需要导出更改作为补丁：
```bash
# 生成补丁文件
git diff main..feat/testnet-upload-env-and-code-delivery-perms > changes.patch

# 应用补丁到其他分支
git apply changes.patch
```

### 关于分支权限

根据当前设置：

1. **分支已创建**: 分支 `feat/testnet-upload-env-and-code-delivery-perms` 已经存在
2. **无需额外权限**: 我在这个预先配置的分支上工作，所有更改都会保存在这个分支上
3. **工作目录清洁**: 当前工作树是干净的（`working tree clean`）

如果遇到权限问题，您可能需要：

#### 给 AI 代理分配权限的方法：

1. **仓库级别的权限**:
   - 在 GitHub/GitLab 等平台上，为运行 AI 的账户添加 `Write` 或 `Maintain` 权限
   - 路径：Repository Settings → Collaborators → Add people

2. **分支保护规则**:
   - 如果分支有保护规则，需要添加例外
   - 路径：Repository Settings → Branches → Branch protection rules
   - 选项：Allow force pushes / Bypass requirements

3. **使用 Personal Access Token (PAT)**:
   - 创建有 `repo` 权限的 PAT
   - 配置 Git 使用 PAT 进行身份验证
   ```bash
   git config --global credential.helper store
   git config --global user.name "Your Name"
   git config --global user.email "your.email@example.com"
   ```

### 替代方案（如果无分支权限）

如果无法创建或推送分支，可以通过以下方式获取代码：

#### 方案 1：导出更改文件
```bash
# 列出所有修改的文件
git diff --name-only > modified_files.txt

# 复制修改的文件到单独目录
mkdir /tmp/changes
git diff --name-only | xargs -I {} cp --parents {} /tmp/changes/
```

#### 方案 2：生成格式化的更改报告
```bash
# 生成详细的更改报告
git diff main > detailed_changes.diff

# 生成统计信息
git diff --stat main > changes_summary.txt
```

#### 方案 3：通过 Pull Request
1. Fork 原始仓库
2. 在 Fork 中创建分支并提交更改
3. 创建 Pull Request 回原仓库
4. 代码审查后合并

### 验证更改

完成后，您可以验证更改：

```bash
# 查看该分支的所有提交
git log --graph --oneline --all

# 查看特定文件的更改历史
git log -p -- TESTNET_UPLOAD_GUIDE.md

# 比较两个分支
git diff main feat/testnet-upload-env-and-code-delivery-perms
```

---

## 总结

1. **环境变量参数**: 最少需要 `--url`, `--key`, `--file` 和 `--indexer`（或 `--node`）四个必需参数
2. **代码获取**: 所有更改都在 `feat/testnet-upload-env-and-code-delivery-perms` 分支上，使用 `git checkout` 和 `git pull` 获取
3. **分支权限**: 当前分支已配置好，如需额外权限可在仓库设置中添加
