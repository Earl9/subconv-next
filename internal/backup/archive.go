package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	FormatVersion               = 1
	MaxArchiveSize        int64 = 16 << 20
	MaxExpandedSize       int64 = 64 << 20
	MaxFileSize           int64 = 16 << 20
	MaxFiles                    = 2000
	serviceCommandTimeout       = 30 * time.Second
)

var safeDataDir = regexp.MustCompile(`^/(?:etc/subconv-next|mnt|overlay)/[A-Za-z0-9._/-]+$`)

type FileRecord struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	FormatVersion int          `json:"format_version"`
	Application   string       `json:"application"`
	AppVersion    string       `json:"app_version"`
	CreatedAt     string       `json:"created_at"`
	Includes      []string     `json:"includes"`
	Files         []FileRecord `json:"files"`
}

type ExportOptions struct {
	ConfigPath string
	DataDir    string
	OutputPath string
	AppVersion string
}

type ExportResult struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
}

type RestoreOptions struct {
	ArchivePath string
	ConfigPath  string
	DataDir     string
	InitScript  string
	LockPath    string
}

type RestoreResult struct {
	RestoredFiles  int  `json:"restored_files"`
	ServiceRunning bool `json:"service_running"`
}

type sourceFile struct {
	archivePath string
	sourcePath  string
	size        int64
}

func ValidateDataDir(value string) error {
	if !safeDataDir.MatchString(value) || strings.Contains(value, "//") {
		return errors.New("数据目录不在允许范围内")
	}
	clean := filepath.Clean(value)
	if clean != value || clean == "/etc/subconv-next" || clean == "/mnt" || clean == "/overlay" {
		return errors.New("数据目录不在允许范围内")
	}
	return nil
}

func Export(opts ExportOptions) (ExportResult, error) {
	if err := ValidateDataDir(opts.DataDir); err != nil {
		return ExportResult{}, err
	}
	files, err := collectSources(opts.ConfigPath, opts.DataDir)
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0700); err != nil {
		return ExportResult{}, fmt.Errorf("无法创建备份目录: %w", err)
	}
	out, err := os.OpenFile(opts.OutputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return ExportResult{}, fmt.Errorf("无法创建备份文件: %w", err)
	}
	ok := false
	defer func() {
		_ = out.Close()
		if !ok {
			_ = os.Remove(opts.OutputPath)
		}
	}()

	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	manifest := Manifest{
		FormatVersion: FormatVersion,
		Application:   "subconv-next",
		AppVersion:    opts.AppVersion,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Includes:      []string{"service_config", "business_data"},
	}
	for _, item := range files {
		record, writeErr := writeSource(tw, item)
		if writeErr != nil {
			return ExportResult{}, writeErr
		}
		manifest.Files = append(manifest.Files, record)
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return ExportResult{}, fmt.Errorf("无法生成备份清单: %w", err)
	}
	if err := tw.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(manifestJSON)), ModTime: time.Now().UTC(), Typeflag: tar.TypeReg}); err != nil {
		return ExportResult{}, fmt.Errorf("无法写入备份清单: %w", err)
	}
	if _, err := tw.Write(manifestJSON); err != nil {
		return ExportResult{}, fmt.Errorf("无法写入备份清单: %w", err)
	}
	if err := tw.Close(); err != nil {
		return ExportResult{}, fmt.Errorf("无法完成备份归档: %w", err)
	}
	if err := gz.Close(); err != nil {
		return ExportResult{}, fmt.Errorf("无法完成备份压缩: %w", err)
	}
	if err := out.Close(); err != nil {
		return ExportResult{}, fmt.Errorf("无法保存备份文件: %w", err)
	}
	info, err := os.Stat(opts.OutputPath)
	if err != nil {
		return ExportResult{}, err
	}
	if info.Size() > MaxArchiveSize {
		return ExportResult{}, errors.New("备份文件超过 16 MiB 限制")
	}
	sum, err := fileSHA256(opts.OutputPath)
	if err != nil {
		return ExportResult{}, err
	}
	ok = true
	return ExportResult{Path: opts.OutputPath, Filename: backupFilename(time.Now()), Size: info.Size(), SHA256: sum}, nil
}

func Inspect(archivePath string) (Manifest, error) {
	return scanArchive(archivePath, nil)
}

func Restore(opts RestoreOptions) (result RestoreResult, retErr error) {
	if err := ValidateDataDir(opts.DataDir); err != nil {
		return result, err
	}
	lockPath := opts.LockPath
	if lockPath == "" {
		lockPath = "/tmp/subconv-next-restore.lock"
	}
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return result, fmt.Errorf("无法创建恢复锁: %w", err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = lock.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return result, errors.New("已有恢复任务正在执行")
		}
		return result, fmt.Errorf("无法锁定恢复任务: %w", err)
	}
	defer func() {
		_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		_ = lock.Close()
	}()
	if err := lock.Truncate(0); err != nil {
		return result, fmt.Errorf("无法初始化恢复锁: %w", err)
	}
	if _, err := lock.Seek(0, io.SeekStart); err != nil {
		return result, fmt.Errorf("无法初始化恢复锁: %w", err)
	}
	_, _ = fmt.Fprintf(lock, "%d\n", os.Getpid())

	manifest, err := Inspect(opts.ArchivePath)
	if err != nil {
		return result, err
	}
	if err := ensureSpace(filepath.Dir(opts.DataDir), manifest); err != nil {
		return result, err
	}

	token, err := randomToken(8)
	if err != nil {
		return result, err
	}
	dataParent := filepath.Dir(opts.DataDir)
	tempData := filepath.Join(dataParent, ".subconv-next-restore-"+token)
	rollbackData := filepath.Join(dataParent, ".subconv-next-rollback-"+token)
	configDir := filepath.Dir(opts.ConfigPath)
	tempConfig := filepath.Join(configDir, ".subconv-next-restore-"+token)
	rollbackConfig := filepath.Join(configDir, ".subconv-next-rollback-"+token)
	if err := os.Mkdir(tempData, 0700); err != nil {
		return result, fmt.Errorf("无法创建恢复临时目录: %w", err)
	}
	defer os.RemoveAll(tempData)
	defer os.Remove(tempConfig)
	defer os.Remove(rollbackConfig)
	defer os.RemoveAll(rollbackData)

	if _, err := scanArchive(opts.ArchivePath, func(name string, mode os.FileMode, reader io.Reader, size int64) error {
		switch {
		case name == "config/subconv-next":
			return writeRegularFile(tempConfig, reader, size, 0600)
		case strings.HasPrefix(name, "data/"):
			rel := strings.TrimPrefix(name, "data/")
			destination := filepath.Join(tempData, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
				return err
			}
			return writeRegularFile(destination, reader, size, 0600)
		default:
			return nil
		}
	}); err != nil {
		return result, err
	}
	configBytes, err := os.ReadFile(tempConfig)
	if err != nil {
		return result, errors.New("备份中缺少服务配置")
	}
	configBytes, err = preserveDataDir(configBytes, opts.DataDir)
	if err != nil {
		return result, err
	}
	if err := os.WriteFile(tempConfig, configBytes, 0600); err != nil {
		return result, fmt.Errorf("无法准备服务配置: %w", err)
	}

	wasRunning := serviceRunning(opts.InitScript)
	if wasRunning {
		if err := serviceAction(opts.InitScript, "stop"); err != nil {
			return result, errors.New("服务无法停止，恢复已取消")
		}
	}

	configReplaced := false
	dataReplaced := false
	defer func() {
		if retErr == nil {
			return
		}
		_ = serviceAction(opts.InitScript, "stop")
		if dataReplaced {
			_ = os.RemoveAll(opts.DataDir)
			if _, err := os.Stat(rollbackData); err == nil {
				_ = os.Rename(rollbackData, opts.DataDir)
			}
		}
		if configReplaced {
			_ = os.Remove(opts.ConfigPath)
			_ = os.Rename(rollbackConfig, opts.ConfigPath)
		}
		if wasRunning {
			_ = serviceAction(opts.InitScript, "start")
		}
	}()

	if err := copyFile(opts.ConfigPath, rollbackConfig, 0600); err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("无法创建配置回滚副本: %w", err)
	}
	if _, err := os.Stat(opts.DataDir); err == nil {
		if err := os.Rename(opts.DataDir, rollbackData); err != nil {
			return result, fmt.Errorf("无法创建数据回滚副本: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("无法检查当前数据目录: %w", err)
	}
	if err := os.Rename(tempData, opts.DataDir); err != nil {
		if _, statErr := os.Stat(rollbackData); statErr == nil {
			_ = os.Rename(rollbackData, opts.DataDir)
		}
		return result, fmt.Errorf("无法替换数据目录: %w", err)
	}
	dataReplaced = true
	if err := os.Chmod(opts.DataDir, 0700); err != nil {
		return result, fmt.Errorf("无法设置数据目录权限: %w", err)
	}
	if err := os.Rename(tempConfig, opts.ConfigPath); err != nil {
		return result, fmt.Errorf("无法替换服务配置: %w", err)
	}
	configReplaced = true
	if err := os.Chmod(opts.ConfigPath, 0644); err != nil {
		return result, fmt.Errorf("无法设置配置权限: %w", err)
	}
	if wasRunning {
		if err := serviceAction(opts.InitScript, "start"); err != nil || !serviceRunning(opts.InitScript) {
			return result, errors.New("恢复后服务启动失败，已回滚")
		}
	}
	_ = os.Remove(rollbackConfig)
	_ = os.RemoveAll(rollbackData)
	return RestoreResult{RestoredFiles: len(manifest.Files), ServiceRunning: wasRunning}, nil
}

func collectSources(configPath, dataDir string) ([]sourceFile, error) {
	configInfo, err := os.Lstat(configPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取服务配置: %w", err)
	}
	if !configInfo.Mode().IsRegular() {
		return nil, errors.New("服务配置不是普通文件")
	}
	files := []sourceFile{{archivePath: "config/subconv-next", sourcePath: configPath, size: configInfo.Size()}}
	var total int64 = configInfo.Size()
	err = filepath.WalkDir(dataDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath == dataDir {
			return nil
		}
		rel, err := filepath.Rel(dataDir, filePath)
		if err != nil {
			return err
		}
		if excludedPath(rel, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if info.Size() > MaxFileSize {
			return fmt.Errorf("业务文件过大: %s", rel)
		}
		total += info.Size()
		if total > MaxExpandedSize {
			return errors.New("业务数据超过 64 MiB 限制")
		}
		files = append(files, sourceFile{archivePath: path.Join("data", filepath.ToSlash(rel)), sourcePath: filePath, size: info.Size()})
		if len(files) > MaxFiles {
			return errors.New("业务文件数量超过限制")
		}
		return nil
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("无法收集业务数据: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].archivePath < files[j].archivePath })
	return files, nil
}

func excludedPath(rel string, directory bool) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts {
		lower := strings.ToLower(part)
		if lower == "logs" || lower == "cache" || lower == "tmp" || lower == "temp" {
			return true
		}
	}
	name := strings.ToLower(parts[len(parts)-1])
	if strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".pid") || strings.HasSuffix(name, ".lock") || strings.HasSuffix(name, ".tmp") {
		return true
	}
	if !directory && (strings.Contains(name, "private-key") || strings.Contains(name, "secret") || strings.Contains(name, "token")) {
		return true
	}
	return false
}

func writeSource(tw *tar.Writer, item sourceFile) (FileRecord, error) {
	file, err := os.Open(item.sourcePath)
	if err != nil {
		return FileRecord{}, fmt.Errorf("无法读取备份文件 %s: %w", item.archivePath, err)
	}
	defer file.Close()
	header := &tar.Header{Name: item.archivePath, Mode: 0600, Size: item.size, ModTime: time.Now().UTC(), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(header); err != nil {
		return FileRecord{}, err
	}
	hash := sha256.New()
	written, err := io.CopyN(io.MultiWriter(tw, hash), file, item.size)
	if err != nil || written != item.size {
		return FileRecord{}, fmt.Errorf("读取备份文件失败: %s", item.archivePath)
	}
	return FileRecord{Path: item.archivePath, Size: item.size, SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}

type archiveSink func(name string, mode os.FileMode, reader io.Reader, size int64) error

func scanArchive(archivePath string, sink archiveSink) (Manifest, error) {
	info, err := os.Lstat(archivePath)
	if err != nil {
		return Manifest{}, errors.New("找不到上传的备份文件")
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > MaxArchiveSize {
		return Manifest{}, errors.New("备份文件大小无效或超过 16 MiB")
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return Manifest{}, errors.New("无法读取备份文件")
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return Manifest{}, errors.New("文件不是有效的 tar.gz 备份")
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	observed := make(map[string]FileRecord)
	var manifestBytes []byte
	var total int64
	for count := 0; ; count++ {
		header, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return Manifest{}, errors.New("备份归档已损坏")
		}
		if count >= MaxFiles+1 {
			return Manifest{}, errors.New("备份文件数量超过限制")
		}
		if err := validateArchiveName(header.Name); err != nil {
			return Manifest{}, err
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != 0 {
			return Manifest{}, fmt.Errorf("备份包含不支持的文件类型: %s", header.Name)
		}
		if header.Size < 0 || header.Size > MaxFileSize {
			return Manifest{}, fmt.Errorf("备份中的文件大小无效: %s", header.Name)
		}
		total += header.Size
		if total > MaxExpandedSize {
			return Manifest{}, errors.New("备份解压后超过 64 MiB 限制")
		}
		if header.Name == "manifest.json" {
			if manifestBytes != nil || header.Size > 1<<20 {
				return Manifest{}, errors.New("备份清单无效")
			}
			manifestBytes, err = io.ReadAll(io.LimitReader(tr, header.Size))
			if err != nil || int64(len(manifestBytes)) != header.Size {
				return Manifest{}, errors.New("备份清单已损坏")
			}
			continue
		}
		if header.Name != "config/subconv-next" && !strings.HasPrefix(header.Name, "data/") {
			return Manifest{}, fmt.Errorf("备份包含未知路径: %s", header.Name)
		}
		if _, exists := observed[header.Name]; exists {
			return Manifest{}, fmt.Errorf("备份包含重复路径: %s", header.Name)
		}
		hash := sha256.New()
		reader := io.TeeReader(io.LimitReader(tr, header.Size), hash)
		if sink != nil {
			if err := sink(header.Name, os.FileMode(header.Mode)&0777, reader, header.Size); err != nil {
				return Manifest{}, fmt.Errorf("无法准备恢复文件 %s: %w", header.Name, err)
			}
		} else if _, err := io.Copy(io.Discard, reader); err != nil {
			return Manifest{}, errors.New("备份归档已损坏")
		}
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return Manifest{}, errors.New("备份归档已损坏")
		}
		observed[header.Name] = FileRecord{Path: header.Name, Size: header.Size, SHA256: hex.EncodeToString(hash.Sum(nil))}
	}
	if manifestBytes == nil {
		return Manifest{}, errors.New("备份中缺少 manifest.json")
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Manifest{}, errors.New("manifest.json 格式无效")
	}
	if manifest.FormatVersion != FormatVersion {
		return Manifest{}, fmt.Errorf("不支持的备份格式版本: %d", manifest.FormatVersion)
	}
	if manifest.Application != "subconv-next" || manifest.AppVersion == "" || manifest.CreatedAt == "" || len(manifest.Includes) == 0 {
		return Manifest{}, errors.New("manifest.json 字段不完整")
	}
	if _, err := time.Parse(time.RFC3339, manifest.CreatedAt); err != nil {
		return Manifest{}, errors.New("manifest.json 创建时间无效")
	}
	includeSet := make(map[string]bool, len(manifest.Includes))
	for _, item := range manifest.Includes {
		includeSet[item] = true
	}
	if !includeSet["service_config"] || !includeSet["business_data"] {
		return Manifest{}, errors.New("manifest.json 包含范围不完整")
	}
	if len(manifest.Files) != len(observed) {
		return Manifest{}, errors.New("备份文件清单与归档内容不一致")
	}
	hasConfig := false
	manifestPaths := make(map[string]struct{}, len(manifest.Files))
	for _, expected := range manifest.Files {
		if _, duplicate := manifestPaths[expected.Path]; duplicate {
			return Manifest{}, fmt.Errorf("备份清单包含重复路径: %s", expected.Path)
		}
		manifestPaths[expected.Path] = struct{}{}
		actual, exists := observed[expected.Path]
		if !exists || actual.Size != expected.Size || !strings.EqualFold(actual.SHA256, expected.SHA256) {
			return Manifest{}, fmt.Errorf("备份文件校验失败: %s", expected.Path)
		}
		if expected.Path == "config/subconv-next" {
			hasConfig = true
		}
	}
	if !hasConfig {
		return Manifest{}, errors.New("备份中缺少服务配置")
	}
	return manifest, nil
}

func validateArchiveName(name string) error {
	if name == "" || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || path.Clean(name) != name || name == "." || strings.HasPrefix(name, "../") || strings.Contains(name, "/../") {
		return fmt.Errorf("备份包含不安全路径: %s", name)
	}
	return nil
}

func writeRegularFile(destination string, reader io.Reader, size int64, mode os.FileMode) error {
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	written, copyErr := io.CopyN(file, reader, size)
	closeErr := file.Close()
	if copyErr != nil || written != size {
		return errors.New("文件内容不完整")
	}
	return closeErr
}

func preserveDataDir(config []byte, dataDir string) ([]byte, error) {
	lines := strings.Split(string(config), "\n")
	sectionStart := -1
	sectionEnd := len(lines)
	for index, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 || fields[0] != "config" {
			continue
		}
		if sectionStart >= 0 {
			sectionEnd = index
			break
		}
		if len(fields) >= 3 && trimUCIWord(fields[1]) == "subconv-next" && trimUCIWord(fields[2]) == "main" {
			sectionStart = index
		}
	}
	if sectionStart < 0 {
		return nil, errors.New("备份中的服务配置无效")
	}

	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:sectionStart+1]...)
	out = append(out, "\toption data_dir '"+dataDir+"'")
	for index := sectionStart + 1; index < sectionEnd; index++ {
		if isUCIOption(lines[index], "data_dir") {
			continue
		}
		out = append(out, lines[index])
	}
	out = append(out, lines[sectionEnd:]...)
	return []byte(strings.Join(out, "\n")), nil
}

func trimUCIWord(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func isUCIOption(line, option string) bool {
	fields := strings.Fields(strings.TrimSpace(line))
	return len(fields) >= 3 && fields[0] == "option" && trimUCIWord(fields[1]) == option
}

func ensureSpace(targetParent string, manifest Manifest) error {
	var required uint64 = 1 << 20
	for _, file := range manifest.Files {
		required += uint64(file.Size) * 2
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(targetParent, &stat); err != nil {
		return fmt.Errorf("无法检查可用空间: %w", err)
	}
	available := stat.Bavail * uint64(stat.Bsize)
	if available < required {
		return errors.New("可用空间不足，恢复已取消")
	}
	return nil
}

func serviceAction(initScript, action string) error {
	if initScript == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), serviceCommandTimeout)
	defer cancel()
	if err := exec.CommandContext(ctx, initScript, action).Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("服务操作 %s 超时", action)
		}
		return err
	}
	return nil
}

func serviceRunning(initScript string) bool {
	if initScript == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), serviceCommandTimeout)
	defer cancel()
	return exec.CommandContext(ctx, initScript, "running").Run() == nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func fileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func randomToken(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("无法生成安全随机标识: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func backupFilename(now time.Time) string {
	return "subconv-next-backup-" + now.Format("20060102-150405") + ".tar.gz"
}
