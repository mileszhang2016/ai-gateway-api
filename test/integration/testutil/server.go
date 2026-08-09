package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ServerManager 测试服务器管理器
type ServerManager struct {
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	ServerURL  string
	DBPath     string
	DataDir    string
	ConfDir    string
	binPath    string
	tmpConfDir string
}

// StartServer 使用项目编译的 ai-gateway-api.exe 作为子进程启动测试服务器
func StartServer() (*ServerManager, error) {
	sm := &ServerManager{}

	// 1. 获取 integration 目录和项目根目录
	testRoot, err := findIntegrationRoot()
	if err != nil {
		return nil, fmt.Errorf("find integration root: %w", err)
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("find project root: %w", err)
	}

	confDir := filepath.Join(testRoot, "conf")
	dataDir := filepath.Join(testRoot, "data")

	sm.ConfDir = confDir
	sm.DataDir = dataDir

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// 2. 生成唯一数据库文件名
	dbPath := filepath.Join(dataDir, fmt.Sprintf("test_ai_gateway_%d.db", os.Getpid()))
	sm.DBPath = dbPath

	// 3. 初始化 SQLite 数据库（执行 DDL）
	ddlPath := filepath.Join(projectRoot, "db_ddl_sqlite.sql")
	if _, err := os.Stat(ddlPath); os.IsNotExist(err) {
		ddlPath = filepath.Join(testRoot, "..", "..", "db_ddl_sqlite.sql")
	}
	if err := InitTestDB(dbPath, ddlPath); err != nil {
		return nil, fmt.Errorf("init test db: %w", err)
	}

	// 3.5 插入默认测试数据
	if err := SeedTestData(dbPath); err != nil {
		return nil, fmt.Errorf("seed test data: %w", err)
	}

	// 4. 查找真实 ai-gateway-api.exe 二进制
	binSrc := filepath.Join(projectRoot, "ai-gateway-api.exe")
	if _, err := os.Stat(binSrc); os.IsNotExist(err) {
		return nil, fmt.Errorf("ai-gateway-api.exe not found at %s, run 'make build' first", binSrc)
	}

	// 复制到 data 目录（避免文件锁冲突）
	sm.binPath = filepath.Join(dataDir, fmt.Sprintf("ai-gateway-api-%d.exe", os.Getpid()))
	if err := copyFile(binSrc, sm.binPath); err != nil {
		return nil, fmt.Errorf("copy binary: %w", err)
	}

	// 5. 生成随机端口并创建临时配置文件
	port, err := getRandomPort()
	if err != nil {
		return nil, fmt.Errorf("get random port: %w", err)
	}

	tmpConfDir, err := createTempConfig(confDir, sm.binPath, dbPath, port)
	if err != nil {
		return nil, fmt.Errorf("create temp config: %w", err)
	}
	sm.tmpConfDir = tmpConfDir

	// 6. 启动 ai-gateway-api 子进程
	ctx, cancel := context.WithCancel(context.Background())
	sm.cancel = cancel

	serverCmd := exec.CommandContext(ctx, sm.binPath,
		"-c", tmpConfDir+string(os.PathSeparator),
		"-sc", "ai_gateway_api.toml",
		"-l", dataDir,
	)
	serverCmd.Dir = testRoot

	// 捕获 stdout 和 stderr
	stdout, err := serverCmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	var stderrBuf bytes.Buffer
	serverCmd.Stderr = &stderrBuf

	if err := serverCmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start ai-gateway-api: %w", err)
	}
	sm.cmd = serverCmd

	// 7. 等待服务器就绪（TCP 拨号检测，带超时）
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	sm.ServerURL = fmt.Sprintf("http://%s", addr)

	// 启动 stdout 消费 goroutine，防止子进程阻塞
	go func() {
		io.Copy(io.Discard, stdout)
	}()

	if err := sm.waitForReady(addr, 10*time.Second); err != nil {
		// 收集 stderr 输出用于调试
		errMsg := err.Error()
		if stderrBuf.Len() > 0 {
			errMsg = fmt.Sprintf("%s\nstderr: %s", errMsg, stderrBuf.String())
		}
		// 检查进程是否已退出
		if sm.cmd.ProcessState != nil {
			errMsg = fmt.Sprintf("%s\nexit code: %d", errMsg, sm.cmd.ProcessState.ExitCode())
		}
		sm.Shutdown()
		return nil, fmt.Errorf("server ready: %s", errMsg)
	}

	// 设置全局客户端
	SetServerURL(sm.ServerURL)

	return sm, nil
}

// Shutdown 停止测试服务器并清理资源
func (sm *ServerManager) Shutdown() {
	// 取消 context，终止子进程
	if sm.cancel != nil {
		sm.cancel()
	}

	// 等待子进程退出
	if sm.cmd != nil && sm.cmd.Process != nil {
		sm.cmd.Wait()
	}

	// 清理复制的二进制文件
	if sm.binPath != "" {
		os.Remove(sm.binPath)
	}

	// 清理临时配置文件
	if sm.tmpConfDir != "" {
		os.RemoveAll(sm.tmpConfDir)
	}

	// 清理数据库文件
	if sm.DBPath != "" {
		CleanupTestDB(sm.DBPath)
	}
}

// waitForReady 等待服务器就绪
func (sm *ServerManager) waitForReady(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server not ready at %s in %v", addr, timeout)
}

// getRandomPort 获取一个随机可用端口
func getRandomPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// createTempConfig 创建临时配置文件（覆盖端口和数据库路径）
func createTempConfig(srcConfDir, binPath, dbPath string, port int) (string, error) {
	// 创建临时配置目录
	tmpDir, err := os.MkdirTemp("", "ai-gateway-test-conf-")
	if err != nil {
		return "", err
	}

	// 复制所有配置文件
	entries, err := os.ReadDir(srcConfDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		src := filepath.Join(srcConfDir, entry.Name())
		dst := filepath.Join(tmpDir, entry.Name())
		if entry.IsDir() {
			if err := copyDir(src, dst); err != nil {
				os.RemoveAll(tmpDir)
				return "", err
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				os.RemoveAll(tmpDir)
				return "", err
			}
		}
	}

	// 修改主配置文件：端口、数据库路径
	confFile := filepath.Join(tmpDir, "ai_gateway_api.toml")
	content, err := os.ReadFile(confFile)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	confStr := string(content)
	// 替换端口
	confStr = strings.Replace(confStr, "ServerPort = 8199", fmt.Sprintf("ServerPort = %d", port), 1)
	// 替换数据库路径（使用正斜杠避免 TOML 转义问题）
	dbPathForTOML := strings.ReplaceAll(dbPath, "\\", "/")
	confStr = strings.Replace(confStr, `DBName  = "./data/test_ai_gateway.db"`, fmt.Sprintf(`DBName  = "%s"`, dbPathForTOML), 1)

	if err := os.WriteFile(confFile, []byte(confStr), 0644); err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return tmpDir, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

// copyDir 复制目录
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// findIntegrationRoot 查找 integration 目录
func findIntegrationRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller info")
	}
	dir := filepath.Dir(filepath.Dir(filename))
	return dir, nil
}

// findProjectRoot 查找 ai-gateway-api 项目根目录
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			content, err := os.ReadFile(goModPath)
			if err == nil && strings.Contains(string(content), "module github.com/infinity-ai-gateway/ai-gateway-api") && !strings.Contains(string(content), "integration") {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found from %s", cwd)
}
