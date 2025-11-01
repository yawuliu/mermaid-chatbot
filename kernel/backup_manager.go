// kernel/backup_manager.go
package kernel

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cleanupOldBackups 清理旧的备份文件
func (pm *PersistenceManager) cleanupOldBackups() {
	if pm.kernel.config == nil || !pm.kernel.config.EnableAutoBackup {
		return
	}

	backupDir := filepath.Join(pm.corpusPath, "backups")
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	// 按修改时间排序
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := files[i].Info()
		infoJ, _ := files[j].Info()
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 删除超过数量限制的旧备份
	maxFiles := pm.kernel.config.MaxBackupFiles
	if len(files) > maxFiles {
		for i := 0; i < len(files)-maxFiles; i++ {
			oldFile := filepath.Join(backupDir, files[i].Name())
			os.Remove(oldFile)
		}
	}
}

// RestoreFromBackup 从备份恢复文件
func (pm *PersistenceManager) RestoreFromBackup(backupFilename string) error {
	backupPath := filepath.Join(pm.corpusPath, "backups", backupFilename)
	originalName := strings.Split(backupFilename, ".backup.")[0]
	originalPath := filepath.Join(pm.corpusPath, originalName)

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(originalPath, backupData, 0644)
}

// ListBackups 列出所有备份文件
func (pm *PersistenceManager) ListBackups() ([]string, error) {
	backupDir := filepath.Join(pm.corpusPath, "backups")
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	var backups []string
	for _, file := range files {
		backups = append(backups, file.Name())
	}

	return backups, nil
}
