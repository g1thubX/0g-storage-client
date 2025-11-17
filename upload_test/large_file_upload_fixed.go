package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/0gfoundation/0g-storage-client/common"
	"github.com/0gfoundation/0g-storage-client/common/blockchain"
	"github.com/0gfoundation/0g-storage-client/core"
	"github.com/0gfoundation/0g-storage-client/indexer"
	"github.com/0gfoundation/0g-storage-client/node"
	"github.com/0gfoundation/0g-storage-client/transfer"
	"github.com/ethereum/go-ethereum/common/hexutil"
	providers "github.com/openweb3/go-rpc-provider/provider_wrapper"
	"github.com/openweb3/web3go"
	"github.com/sirupsen/logrus"
)

func main() {
	// 设置日志级别
	logrus.SetLevel(logrus.InfoLevel)

	// 环境变量记录
	fmt.Println("=== 使用的环境变量 ===")
	fmt.Printf("Testnet_RPC: %s\n", os.Getenv("Testnet_RPC"))
	fmt.Printf("Testnet_Indexer: %s\n", os.Getenv("Testnet_Indexer"))
	fmt.Printf("Testnet_ANKR_RPC: %s\n", os.Getenv("Testnet_ANKR_RPC"))
	fmt.Printf("PRIVATE_KEY: [已设置但隐藏]\n")
	fmt.Println("========================")

	// 创建4GB测试文件
	filePath := "4gb_test_file.bin"
	if err := createLargeFile(filePath, 4*1024*1024*1024); err != nil {
		log.Fatalf("创建测试文件失败: %v", err)
	}
	fmt.Printf("成功创建4GB测试文件: %s\n", filePath)

	// 切分文件为10个400MB文件
	fragmentDir := "fragments"
	if err := os.MkdirAll(fragmentDir, 0755); err != nil {
		log.Fatalf("创建分片目录失败: %v", err)
	}

	fragmentFiles, err := splitFile(filePath, fragmentDir, 10)
	if err != nil {
		log.Fatalf("文件切分失败: %v", err)
	}
	fmt.Printf("成功切分为%d个文件:\n", len(fragmentFiles))
	for i, file := range fragmentFiles {
		fmt.Printf("  分片%d: %s\n", i+1, file)
	}

	// 上传所有分片
	ctx := context.Background()
	uploadedRoots := make([]string, 0, len(fragmentFiles))

	// 从环境变量获取配置
	rpcURL := os.Getenv("Testnet_RPC")
	if rpcURL == "" {
		rpcURL = os.Getenv("Testnet_ANKR_RPC")
	}
	privateKey := os.Getenv("PRIVATE_KEY")
	indexerURL := os.Getenv("Testnet_Indexer")

	if rpcURL == "" || privateKey == "" || indexerURL == "" {
		log.Fatalf("缺少必需的环境变量: Testnet_RPC/Testnet_ANKR_RPC, PRIVATE_KEY, Testnet_Indexer")
	}

	// 创建web3客户端
	w3client := blockchain.MustNewWeb3(rpcURL, privateKey, providers.Option{})
	defer w3client.Close()

	// 创建indexer客户端
	indexerClient, err := indexer.NewClient(indexerURL, indexer.IndexerClientOption{
		ProviderOption: providers.Option{},
		LogOption:      common.LogOption{Logger: logrus.StandardLogger()},
	})
	if err != nil {
		log.Fatalf("创建indexer客户端失败: %v", err)
	}
	defer indexerClient.Close()

	// 上传配置
	fragmentSize := int64(400 * 1024 * 1024) // 400MB
	opt := transfer.UploadOption{
		Tags:             hexutil.MustDecode("0x"),
		FinalityRequired: transfer.FileFinalized, // 等待文件最终确认
		TaskSize:         10,
		ExpectedReplica:  1,
		SkipTx:           false, // 不跳过交易，确保完整上传
		Fee:              nil,
		Nonce:            nil,
		MaxGasPrice:      nil,
		NRetries:         3,
		Step:             15,
		Method:           "min",
		FullTrusted:      true,
	}

	fmt.Println("\n=== 开始上传分片 ===")
	for i, fragmentFile := range fragmentFiles {
		fmt.Printf("上传分片 %d/%d: %s\n", i+1, len(fragmentFiles), fragmentFile)

		file, err := core.Open(fragmentFile)
		if err != nil {
			log.Printf("打开文件失败 %s: %v", fragmentFile, err)
			continue
		}
		defer file.Close()

		uploader, closer, err := newIndexerUploader(ctx, file.NumSegments(), w3client, indexerClient, opt)
		if err != nil {
			log.Printf("创建上传器失败 %s: %v", fragmentFile, err)
			file.Close()
			continue
		}
		defer closer()

		_, roots, err := uploader.SplitableUpload(ctx, file, fragmentSize, opt)
		if err != nil {
			log.Printf("上传失败 %s: %v", fragmentFile, err)
			continue
		}

		if len(roots) > 0 {
			uploadedRoots = append(uploadedRoots, roots[0].String())
			fmt.Printf("  上传成功，root: %s\n", roots[0].String())
		}
	}

	fmt.Printf("\n=== 上传完成 ===")
	fmt.Printf("成功上传 %d/%d 个分片\n", len(uploadedRoots), len(fragmentFiles))
	for i, root := range uploadedRoots {
		fmt.Printf("分片%d root: %s\n", i+1, root)
	}

	// 下载测试
	fmt.Println("\n=== 开始下载测试 ===")
	downloadDir := "downloads"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("创建下载目录失败: %v", err)
	}

	for i, root := range uploadedRoots {
		downloadFile := filepath.Join(downloadDir, fmt.Sprintf("downloaded_fragment_%d.bin", i+1))
		fmt.Printf("下载分片 %d/%d (root: %s) 到 %s\n", i+1, len(uploadedRoots), root, downloadFile)

		if err := downloadFileFromRoot(ctx, w3client, indexerClient, root, downloadFile); err != nil {
			log.Printf("下载失败 %s: %v", root, err)
			continue
		}

		// 验证文件大小
		if info, err := os.Stat(downloadFile); err == nil {
			fmt.Printf("  下载成功，文件大小: %d bytes\n", info.Size())
		}
	}

	fmt.Println("\n=== 所有任务完成 ===")
}

func createLargeFile(filePath string, size int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用随机数据填充文件
	buffer := make([]byte, 1024*1024) // 1MB buffer
	written := int64(0)

	for written < size {
		toWrite := int64(len(buffer))
		if written+toWrite > size {
			toWrite = size - written
		}

		if _, err := rand.Read(buffer[:toWrite]); err != nil {
			return err
		}

		n, err := file.Write(buffer[:toWrite])
		if err != nil {
			return err
		}
		written += int64(n)
	}

	return nil
}

func splitFile(filePath string, outputDir string, numFragments int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fragmentSize := fileInfo.Size() / int64(numFragments)
	fragmentFiles := make([]string, 0, numFragments)

	for i := 0; i < numFragments; i++ {
		fragmentFile := filepath.Join(outputDir, fmt.Sprintf("fragment_%d.bin", i+1))
		fragmentFiles = append(fragmentFiles, fragmentFile)

		outFile, err := os.Create(fragmentFile)
		if err != nil {
			return nil, err
		}
		defer outFile.Close()

		// 处理最后一个分片可能的大小差异
		var currentFragmentSize int64 = fragmentSize
		if i == numFragments-1 {
			currentFragmentSize = fileInfo.Size() - int64(i)*fragmentSize
		}

		copied, err := io.CopyN(outFile, file, currentFragmentSize)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if copied != currentFragmentSize {
			return nil, fmt.Errorf("复制字节数不匹配: 期望 %d, 实际 %d", currentFragmentSize, copied)
		}
	}

	return fragmentFiles, nil
}

func newIndexerUploader(ctx context.Context, segNum uint64, w3client *web3go.Client, indexerClient *indexer.Client, opt transfer.UploadOption) (*transfer.Uploader, func(), error) {
	up, err := indexerClient.NewUploaderFromIndexerNodes(ctx, segNum, w3client, opt.ExpectedReplica, nil, opt.Method, opt.FullTrusted)
	if err != nil {
		return nil, nil, err
	}

	return up, indexerClient.Close, nil
}

func downloadFileFromRoot(ctx context.Context, w3client *web3go.Client, indexerClient *indexer.Client, root string, outputPath string) error {
	// 从indexer获取可用的存储节点
	shardedNodes, err := indexerClient.GetShardedNodes(ctx)
	if err != nil {
		return fmt.Errorf("获取存储节点失败: %v", err)
	}

	// 合并可信节点和发现节点
	allNodes := append(shardedNodes.Trusted, shardedNodes.Discovered...)
	if len(allNodes) == 0 {
		return fmt.Errorf("没有可用的存储节点")
	}

	// 创建存储节点客户端
	clients := make([]*node.ZgsClient, 0, len(allNodes))
	for _, shardedNode := range allNodes {
		zgsClient, err := node.NewZgsClient(shardedNode.URL, providers.Option{})
		if err != nil {
			logrus.WithError(err).Warnf("创建节点客户端失败: %s", shardedNode.URL)
			continue
		}
		clients = append(clients, zgsClient)
	}

	if len(clients) == 0 {
		return fmt.Errorf("无法创建任何存储节点客户端")
	}

	// 创建下载器
	downloader, err := transfer.NewDownloader(clients, common.LogOption{Logger: logrus.StandardLogger()})
	if err != nil {
		return fmt.Errorf("创建下载器失败: %v", err)
	}

	// 执行下载
	err = downloader.Download(ctx, root, outputPath, false) // 不需要proof验证
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}

	// 关闭所有客户端
	for _, client := range clients {
		client.Close()
	}

	return nil
}