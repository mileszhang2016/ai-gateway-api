package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/yf-networks/ai-gateway-api/stateful"
)

// InitTestDB 初始化 SQLite 测试数据库
// 创建数据库文件并执行 DDL
func InitTestDB(dbPath, ddlPath string) error {
	// 确保数据目录存在
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// 删除旧数据库文件（包括 WAL 和 SHM）
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	// 打开数据库连接（使用自定义驱动剥离 FOR UPDATE）
	db, err := sql.Open("sqlite-strip", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}
	defer db.Close()

	// 读取并执行 DDL 文件
	ddlContent, err := os.ReadFile(ddlPath)
	if err != nil {
		return fmt.Errorf("read ddl file: %w", err)
	}

	// 规范化换行符（Windows \r\n -> \n），然后按 ;\n 分割
	// 不使用 ; 单独分割，避免触发器中 BEGIN ... END; 内的分号被错误拆分
	normalized := strings.ReplaceAll(string(ddlContent), "\r\n", "\n")
	statements := strings.Split(normalized, ";\n")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		// 移除注释行（以 -- 开头的行），保留非注释内容
		lines := strings.Split(stmt, "\n")
		var cleanLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				cleanLines = append(cleanLines, line)
			}
		}
		stmt = strings.TrimSpace(strings.Join(cleanLines, "\n"))
		if stmt == "" {
			continue
		}
		// 补充末尾的分号（被 split 消费掉了）
		execSQL := stmt + ";"
		if _, err := db.Exec(execSQL); err != nil {
			return fmt.Errorf("exec ddl: %s: %w", truncateSQL(stmt, 80), err)
		}
	}

	return nil
}

// CleanupTestDB 清理测试数据库文件
func CleanupTestDB(dbPath string) {
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")
}

func truncateSQL(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
