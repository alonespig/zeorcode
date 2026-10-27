package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/repository"
	"github.com/alonespig/zeorcode/backend/pkg/errcode"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	ProblemArchiveFormat  = "zeorcode-problems"
	ProblemArchiveVersion = 1

	MaxProblemArchiveUploadSize int64 = 2 << 30
	maxArchiveManifestSize      int64 = 32 << 20
	maxArchiveUncompressedSize  int64 = 20 << 30
	maxArchiveFileSize          int64 = 256 << 20
	maxArchiveProblems                = 10_000
	maxArchiveFiles                   = 100_000
	maxArchiveFilesPerProblem         = 1_000
	maxArchiveTagsPerProblem          = 100
	maxArchiveSamplesPerProblem       = 1_000
)

type ProblemArchiveManifest struct {
	Format     string                  `json:"format"`
	Version    int                     `json:"version"`
	ExportedAt string                  `json:"exportedAt"`
	Problems   []ProblemArchiveProblem `json:"problems"`
}

type ProblemArchiveProblem struct {
	DisplayID       string                 `json:"displayId"`
	Name            string                 `json:"name"`
	Difficulty      int                    `json:"difficulty"`
	TimeLimit       int                    `json:"timeLimit"`
	MemoryLimit     int                    `json:"memoryLimit"`
	Description     string                 `json:"description"`
	InputFormat     string                 `json:"inputFormat"`
	OutputFormat    string                 `json:"outputFormat"`
	Hint            string                 `json:"hint"`
	OJ              string                 `json:"oj"`
	RemoteProblemID string                 `json:"remoteProblemId"`
	Hidden          bool                   `json:"hidden"`
	CreatedAt       string                 `json:"createdAt"`
	Tags            []string               `json:"tags"`
	Samples         []ProblemArchiveSample `json:"samples"`
	TestData        []ProblemArchiveFile   `json:"testData"`
}

type ProblemArchiveSample struct {
	Input   string `json:"input"`
	Output  string `json:"output"`
	Explain string `json:"explain"`
}

type ProblemArchiveFile struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type ProblemArchiveImportResult struct {
	Total       int      `json:"total"`
	Imported    int      `json:"imported"`
	Skipped     int      `json:"skipped"`
	ImportedIDs []string `json:"importedIds"`
	SkippedIDs  []string `json:"skippedIds"`
}

type problemArchiveExportRecord struct {
	problemID int64
	index     int
}

// ProblemArchiveExport contains database metadata prepared before response
// headers are written. WriteTo then streams test data directly to the client.
type ProblemArchiveExport struct {
	dataDir  string
	manifest ProblemArchiveManifest
	records  []problemArchiveExportRecord
}

type problemArchiveTagRow struct {
	ProblemID int64
	Name      string
}

func archiveDataDir() string {
	dataDir := strings.TrimSpace(viper.GetString("judge.data_dir"))
	if dataDir == "" {
		dataDir = "ojdata/problems"
	}
	return filepath.Clean(dataDir)
}

// PrepareProblemArchive loads the metadata needed to export all problems.
func (s *ProblemService) PrepareProblemArchive(ctx context.Context) (*ProblemArchiveExport, error) {
	var problems []model.Problem
	if err := s.db.WithContext(ctx).Order("id ASC").Find(&problems).Error; err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	if len(problems) > maxArchiveProblems {
		return nil, errcode.ErrInvalidParams.WithMsg("题目数量超过单次导出上限")
	}

	problemIDs := make([]int64, 0, len(problems))
	for _, problem := range problems {
		problemIDs = append(problemIDs, problem.ID)
	}

	samplesByProblem := make(map[int64][]ProblemArchiveSample, len(problems))
	tagsByProblem := make(map[int64][]string, len(problems))
	if len(problemIDs) > 0 {
		var samples []model.ProblemSamples
		if err := s.db.WithContext(ctx).
			Where("problem_id IN ?", problemIDs).
			Order("problem_id ASC, id ASC").
			Find(&samples).Error; err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
		for _, sample := range samples {
			samplesByProblem[sample.ProblemID] = append(samplesByProblem[sample.ProblemID], ProblemArchiveSample{
				Input: sample.Input, Output: sample.Output, Explain: sample.Explain,
			})
		}

		var tags []problemArchiveTagRow
		if err := s.db.WithContext(ctx).Table("problem_tags").
			Select("problem_tags.problem_id, tags.name").
			Joins("JOIN tags ON tags.id = problem_tags.tag_id").
			Where("problem_tags.problem_id IN ?", problemIDs).
			Order("problem_tags.problem_id ASC, tags.name ASC").
			Scan(&tags).Error; err != nil {
			return nil, errcode.ErrDatabase.Wrap(err)
		}
		for _, tag := range tags {
			tagsByProblem[tag.ProblemID] = append(tagsByProblem[tag.ProblemID], tag.Name)
		}
	}

	archive := &ProblemArchiveExport{
		dataDir: archiveDataDir(),
		manifest: ProblemArchiveManifest{
			Format:     ProblemArchiveFormat,
			Version:    ProblemArchiveVersion,
			ExportedAt: time.Now().UTC().Format(time.RFC3339Nano),
			Problems:   make([]ProblemArchiveProblem, 0, len(problems)),
		},
		records: make([]problemArchiveExportRecord, 0, len(problems)),
	}
	for i, problem := range problems {
		archive.manifest.Problems = append(archive.manifest.Problems, ProblemArchiveProblem{
			DisplayID:       problem.DisplayID,
			Name:            problem.Name,
			Difficulty:      problem.Difficulty,
			TimeLimit:       problem.TimeLimit,
			MemoryLimit:     problem.MemoryLimit,
			Description:     problem.Description,
			InputFormat:     problem.InputFormat,
			OutputFormat:    problem.OutputFormat,
			Hint:            problem.Hint,
			OJ:              problem.OJ,
			RemoteProblemID: problem.RemoteProblemID,
			Hidden:          problem.Hidden,
			CreatedAt:       problem.CreatedAt.UTC().Format(time.RFC3339Nano),
			Tags:            nonNilStrings(tagsByProblem[problem.ID]),
			Samples:         nonNilSamples(samplesByProblem[problem.ID]),
			TestData:        make([]ProblemArchiveFile, 0),
		})
		archive.records = append(archive.records, problemArchiveExportRecord{problemID: problem.ID, index: i})
	}
	return archive, nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func nonNilSamples(values []ProblemArchiveSample) []ProblemArchiveSample {
	if values == nil {
		return []ProblemArchiveSample{}
	}
	return values
}

func archiveProblemPrefix(index int) string {
	return fmt.Sprintf("data/%06d", index+1)
}

func isArchiveTestDataName(name string) bool {
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return false
	}
	return strings.HasSuffix(name, ".in") || strings.HasSuffix(name, ".out")
}

// WriteTo streams the prepared archive. Call it at most once.
func (e *ProblemArchiveExport) WriteTo(ctx context.Context, dst io.Writer) error {
	if err := os.MkdirAll(e.dataDir, 0o750); err != nil {
		return fmt.Errorf("prepare test data root: %w", err)
	}
	root, err := os.OpenRoot(e.dataDir)
	if err != nil {
		return fmt.Errorf("open test data root: %w", err)
	}
	defer root.Close()

	zw := zip.NewWriter(dst)
	closeWithError := func(err error) error {
		_ = zw.Close()
		return err
	}
	fileCount := 0
	var totalSize int64

	for _, record := range e.records {
		if err := ctx.Err(); err != nil {
			return closeWithError(err)
		}
		dirName := strconv.FormatInt(record.problemID, 10)
		dir, err := root.Open(dirName)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return closeWithError(fmt.Errorf("open test data for problem %d: %w", record.problemID, err))
		}
		entries, readErr := dir.ReadDir(-1)
		closeErr := dir.Close()
		if readErr != nil {
			return closeWithError(fmt.Errorf("list test data for problem %d: %w", record.problemID, readErr))
		}
		if closeErr != nil {
			return closeWithError(fmt.Errorf("close test data directory for problem %d: %w", record.problemID, closeErr))
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

		problemFileCount := 0
		for _, entry := range entries {
			if !isArchiveTestDataName(entry.Name()) {
				continue
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() {
				if err != nil {
					return closeWithError(fmt.Errorf("inspect test data %q: %w", entry.Name(), err))
				}
				continue
			}
			if info.Size() > maxArchiveFileSize {
				return closeWithError(fmt.Errorf("test data %q exceeds archive file limit", entry.Name()))
			}
			problemFileCount++
			fileCount++
			totalSize += info.Size()
			if problemFileCount > maxArchiveFilesPerProblem || fileCount > maxArchiveFiles {
				return closeWithError(fmt.Errorf("test data file count exceeds archive limit"))
			}
			if totalSize > maxArchiveUncompressedSize {
				return closeWithError(fmt.Errorf("test data size exceeds archive limit"))
			}

			archiveName := path.Join(archiveProblemPrefix(record.index), entry.Name())
			header := &zip.FileHeader{Name: archiveName, Method: zip.Deflate, Modified: info.ModTime()}
			header.SetMode(0o600)
			writer, err := zw.CreateHeader(header)
			if err != nil {
				return closeWithError(fmt.Errorf("create archive entry %q: %w", archiveName, err))
			}
			file, err := root.Open(path.Join(dirName, entry.Name()))
			if err != nil {
				return closeWithError(fmt.Errorf("open test data %q: %w", entry.Name(), err))
			}
			hash := sha256.New()
			size, copyErr := io.Copy(io.MultiWriter(writer, hash), file)
			fileCloseErr := file.Close()
			if copyErr != nil {
				return closeWithError(fmt.Errorf("archive test data %q: %w", entry.Name(), copyErr))
			}
			if fileCloseErr != nil {
				return closeWithError(fmt.Errorf("close test data %q: %w", entry.Name(), fileCloseErr))
			}
			e.manifest.Problems[record.index].TestData = append(
				e.manifest.Problems[record.index].TestData,
				ProblemArchiveFile{Name: entry.Name(), Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))},
			)
		}
	}

	manifestWriter, err := zw.CreateHeader(&zip.FileHeader{Name: "manifest.json", Method: zip.Deflate})
	if err != nil {
		return closeWithError(fmt.Errorf("create archive manifest: %w", err))
	}
	encoder := json.NewEncoder(manifestWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(e.manifest); err != nil {
		return closeWithError(fmt.Errorf("write archive manifest: %w", err))
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("finish problem archive: %w", err)
	}
	return nil
}

type parsedProblemArchive struct {
	manifest ProblemArchiveManifest
	entries  map[string]*zip.File
}

func invalidArchive(message string, args ...any) error {
	return errcode.ErrInvalidParams.WithMsg(fmt.Sprintf(message, args...))
}

func readProblemArchive(zr *zip.Reader) (*parsedProblemArchive, error) {
	if len(zr.File) == 0 || len(zr.File) > maxArchiveFiles+1 {
		return nil, invalidArchive("归档文件数量不合法")
	}
	entries := make(map[string]*zip.File, len(zr.File))
	var total uint64
	for _, entry := range zr.File {
		if entry.FileInfo().IsDir() || !validArchivePath(entry.Name) {
			return nil, invalidArchive("归档包含非法路径：%s", entry.Name)
		}
		if entry.Mode()&os.ModeType != 0 {
			return nil, invalidArchive("归档包含不支持的特殊文件：%s", entry.Name)
		}
		if _, exists := entries[entry.Name]; exists {
			return nil, invalidArchive("归档包含重复文件：%s", entry.Name)
		}
		if entry.UncompressedSize64 > uint64(maxArchiveFileSize) && entry.Name != "manifest.json" {
			return nil, invalidArchive("归档中的单个测试文件过大：%s", entry.Name)
		}
		if ^uint64(0)-total < entry.UncompressedSize64 {
			return nil, invalidArchive("归档解压大小溢出")
		}
		total += entry.UncompressedSize64
		if total > uint64(maxArchiveUncompressedSize) {
			return nil, invalidArchive("归档解压后总大小超过 20GB")
		}
		entries[entry.Name] = entry
	}

	manifestEntry, ok := entries["manifest.json"]
	if !ok || manifestEntry.UncompressedSize64 > uint64(maxArchiveManifestSize) {
		return nil, invalidArchive("归档缺少有效的 manifest.json")
	}
	reader, err := manifestEntry.Open()
	if err != nil {
		return nil, invalidArchive("无法读取归档清单")
	}
	defer reader.Close()
	manifestBytes, err := io.ReadAll(io.LimitReader(reader, maxArchiveManifestSize+1))
	if err != nil || int64(len(manifestBytes)) > maxArchiveManifestSize {
		return nil, invalidArchive("归档清单过大或已损坏")
	}
	var manifest ProblemArchiveManifest
	decoder := json.NewDecoder(strings.NewReader(string(manifestBytes)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, invalidArchive("归档清单格式错误：%v", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, invalidArchive("归档清单包含多余内容")
	}
	if err := validateProblemArchiveManifest(&manifest, entries); err != nil {
		return nil, err
	}
	return &parsedProblemArchive{manifest: manifest, entries: entries}, nil
}

func validArchivePath(name string) bool {
	return name != "" && !strings.Contains(name, `\`) && fs.ValidPath(name) && path.Clean(name) == name
}

func validateProblemArchiveManifest(manifest *ProblemArchiveManifest, entries map[string]*zip.File) error {
	if manifest.Format != ProblemArchiveFormat || manifest.Version != ProblemArchiveVersion {
		return invalidArchive("不支持的题目归档格式或版本")
	}
	if len(manifest.Problems) > maxArchiveProblems {
		return invalidArchive("归档中的题目数量不合法")
	}
	allowedEntries := map[string]struct{}{"manifest.json": {}}
	displayIDs := make(map[string]struct{}, len(manifest.Problems))
	fileCount := 0
	for index := range manifest.Problems {
		problem := &manifest.Problems[index]
		if err := validateArchivedProblem(problem); err != nil {
			return err
		}
		if _, exists := displayIDs[problem.DisplayID]; exists {
			return invalidArchive("归档中存在重复题号：%s", problem.DisplayID)
		}
		displayIDs[problem.DisplayID] = struct{}{}
		if len(problem.TestData) > maxArchiveFilesPerProblem {
			return invalidArchive("题目 %s 的测试文件过多", problem.DisplayID)
		}
		seenFiles := make(map[string]struct{}, len(problem.TestData))
		for _, file := range problem.TestData {
			if !isArchiveTestDataName(file.Name) || file.Size < 0 || file.Size > maxArchiveFileSize {
				return invalidArchive("题目 %s 包含非法测试文件", problem.DisplayID)
			}
			if _, exists := seenFiles[file.Name]; exists {
				return invalidArchive("题目 %s 包含重复测试文件：%s", problem.DisplayID, file.Name)
			}
			if !validSHA256(file.SHA256) {
				return invalidArchive("题目 %s 的测试文件校验值无效", problem.DisplayID)
			}
			entryName := path.Join(archiveProblemPrefix(index), file.Name)
			entry, exists := entries[entryName]
			if !exists || int64(entry.UncompressedSize64) != file.Size {
				return invalidArchive("题目 %s 的测试文件缺失或大小不符：%s", problem.DisplayID, file.Name)
			}
			seenFiles[file.Name] = struct{}{}
			allowedEntries[entryName] = struct{}{}
			fileCount++
		}
	}
	if fileCount > maxArchiveFiles {
		return invalidArchive("归档中的测试文件数量过多")
	}
	for name := range entries {
		if _, allowed := allowedEntries[name]; !allowed {
			return invalidArchive("归档包含未在清单中声明的文件：%s", name)
		}
	}
	return nil
}

func validateArchivedProblem(problem *ProblemArchiveProblem) error {
	problem.DisplayID = strings.TrimSpace(problem.DisplayID)
	problem.Name = strings.TrimSpace(problem.Name)
	if problem.DisplayID == "" || utf8.RuneCountInString(problem.DisplayID) > 32 {
		return invalidArchive("归档包含非法题号")
	}
	if problem.Name == "" {
		return invalidArchive("题目 %s 的名称不能为空", problem.DisplayID)
	}
	if problem.Difficulty < 1 || problem.Difficulty > 3 || problem.TimeLimit <= 0 || problem.MemoryLimit <= 0 {
		return invalidArchive("题目 %s 的难度或资源限制无效", problem.DisplayID)
	}
	if utf8.RuneCountInString(problem.OJ) > 20 || utf8.RuneCountInString(problem.RemoteProblemID) > 50 {
		return invalidArchive("题目 %s 的远程题信息过长", problem.DisplayID)
	}
	if _, err := time.Parse(time.RFC3339Nano, problem.CreatedAt); err != nil {
		return invalidArchive("题目 %s 的创建时间无效", problem.DisplayID)
	}
	if len(problem.Tags) > maxArchiveTagsPerProblem {
		return invalidArchive("题目 %s 的标签数量过多", problem.DisplayID)
	}
	if len(problem.Samples) > maxArchiveSamplesPerProblem {
		return invalidArchive("题目 %s 的样例数量过多", problem.DisplayID)
	}
	tagSet := make(map[string]struct{}, len(problem.Tags))
	normalizedTags := make([]string, 0, len(problem.Tags))
	for _, value := range problem.Tags {
		tag := strings.TrimSpace(value)
		if tag == "" || utf8.RuneCountInString(tag) > 64 {
			return invalidArchive("题目 %s 包含非法标签", problem.DisplayID)
		}
		if _, exists := tagSet[tag]; exists {
			continue
		}
		tagSet[tag] = struct{}{}
		normalizedTags = append(normalizedTags, tag)
	}
	problem.Tags = normalizedTags
	return nil
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

// ImportProblemArchive validates the complete ZIP before making changes. Existing
// display IDs are skipped; all new problems are committed atomically.
func (s *ProblemService) ImportProblemArchive(ctx context.Context, zr *zip.Reader) (*ProblemArchiveImportResult, error) {
	parsed, err := readProblemArchive(zr)
	if err != nil {
		return nil, err
	}
	result := &ProblemArchiveImportResult{
		Total:       len(parsed.manifest.Problems),
		ImportedIDs: []string{},
		SkippedIDs:  []string{},
	}
	if len(parsed.manifest.Problems) == 0 {
		return result, nil
	}
	displayIDs := make([]string, 0, len(parsed.manifest.Problems))
	for _, problem := range parsed.manifest.Problems {
		displayIDs = append(displayIDs, problem.DisplayID)
	}
	var existingIDs []string
	if err := s.db.WithContext(ctx).Model(&model.Problem{}).
		Where("display_id IN ?", displayIDs).
		Pluck("display_id", &existingIDs).Error; err != nil {
		return nil, errcode.ErrDatabase.Wrap(err)
	}
	existing := make(map[string]struct{}, len(existingIDs))
	for _, displayID := range existingIDs {
		existing[displayID] = struct{}{}
	}
	pending := make([]int, 0, len(parsed.manifest.Problems))
	for index, problem := range parsed.manifest.Problems {
		if _, found := existing[problem.DisplayID]; found {
			result.SkippedIDs = append(result.SkippedIDs, problem.DisplayID)
			continue
		}
		pending = append(pending, index)
	}
	result.Skipped = len(result.SkippedIDs)
	if len(pending) == 0 {
		return result, nil
	}

	dataDir := archiveDataDir()
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, errcode.ErrInternal.WithMsg("创建题目数据目录失败").Wrap(err)
	}
	stagingDir, err := os.MkdirTemp(dataDir, ".archive-import-*")
	if err != nil {
		return nil, errcode.ErrInternal.WithMsg("创建导入临时目录失败").Wrap(err)
	}
	defer os.RemoveAll(stagingDir)
	for _, index := range pending {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		problemDir := filepath.Join(stagingDir, fmt.Sprintf("%06d", index+1))
		if err := os.Mkdir(problemDir, 0o750); err != nil {
			return nil, errcode.ErrInternal.Wrap(err)
		}
		if err := extractArchivedTestData(parsed, index, problemDir); err != nil {
			return nil, err
		}
	}

	finalDirs := make([]string, 0, len(pending))
	importedIDs := make([]string, 0, len(pending))
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tagIDs := make(map[string]int64)
		for _, index := range pending {
			archived := parsed.manifest.Problems[index]
			createdAt, _ := time.Parse(time.RFC3339Nano, archived.CreatedAt)
			problem := model.Problem{
				DisplayID: archived.DisplayID, Name: archived.Name, Difficulty: archived.Difficulty,
				TimeLimit: archived.TimeLimit, MemoryLimit: archived.MemoryLimit,
				Description: archived.Description, InputFormat: archived.InputFormat,
				OutputFormat: archived.OutputFormat, Hint: archived.Hint, OJ: archived.OJ,
				RemoteProblemID: archived.RemoteProblemID, Hidden: archived.Hidden, CreatedAt: createdAt,
			}
			if err := tx.Create(&problem).Error; err != nil {
				return err
			}
			for _, tagName := range archived.Tags {
				tagID, ok := tagIDs[tagName]
				if !ok {
					var tag model.Tag
					err := tx.Where("name = ?", tagName).First(&tag).Error
					if errors.Is(err, gorm.ErrRecordNotFound) {
						tag = model.Tag{Name: tagName}
						err = tx.Create(&tag).Error
					}
					if err != nil {
						return err
					}
					tagID = tag.ID
					tagIDs[tagName] = tagID
				}
				if err := tx.Create(&model.ProblemTag{ProblemID: problem.ID, TagID: tagID}).Error; err != nil {
					return err
				}
			}
			if len(archived.Samples) > 0 {
				samples := make([]model.ProblemSamples, 0, len(archived.Samples))
				for _, sample := range archived.Samples {
					samples = append(samples, model.ProblemSamples{
						ProblemID: problem.ID, Input: sample.Input, Output: sample.Output, Explain: sample.Explain,
					})
				}
				if err := tx.Create(&samples).Error; err != nil {
					return err
				}
			}

			sourceDir := filepath.Join(stagingDir, fmt.Sprintf("%06d", index+1))
			finalDir := filepath.Join(dataDir, strconv.FormatInt(problem.ID, 10))
			if _, err := os.Stat(finalDir); err == nil || !errors.Is(err, os.ErrNotExist) {
				return errcode.ErrInternal.WithMsg("导入题目的测试数据目录已存在")
			}
			if err := os.Rename(sourceDir, finalDir); err != nil {
				return errcode.ErrInternal.WithMsg("移动导入的测试数据失败").Wrap(err)
			}
			finalDirs = append(finalDirs, finalDir)
			importedIDs = append(importedIDs, problem.DisplayID)
		}
		return nil
	})
	if err != nil {
		for _, dir := range finalDirs {
			_ = os.RemoveAll(dir)
		}
		var appErr *errcode.AppErr
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, errcode.ErrDatabase.WithMsg("批量导入题目失败，未写入任何题目").Wrap(err)
	}
	for _, dir := range finalDirs {
		_ = repository.RebuildInfo(dir)
	}
	result.ImportedIDs = importedIDs
	result.Imported = len(importedIDs)
	return result, nil
}

func extractArchivedTestData(parsed *parsedProblemArchive, index int, targetDir string) error {
	problem := parsed.manifest.Problems[index]
	for _, archivedFile := range problem.TestData {
		entryName := path.Join(archiveProblemPrefix(index), archivedFile.Name)
		entry := parsed.entries[entryName]
		reader, err := entry.Open()
		if err != nil {
			return invalidArchive("无法读取题目 %s 的测试文件：%s", problem.DisplayID, archivedFile.Name)
		}
		destination, err := os.OpenFile(filepath.Join(targetDir, archivedFile.Name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			reader.Close()
			return errcode.ErrInternal.Wrap(err)
		}
		hash := sha256.New()
		written, copyErr := io.Copy(io.MultiWriter(destination, hash), io.LimitReader(reader, maxArchiveFileSize+1))
		closeErr := destination.Close()
		readerCloseErr := reader.Close()
		if copyErr != nil || closeErr != nil || readerCloseErr != nil {
			return invalidArchive("题目 %s 的测试文件已损坏：%s", problem.DisplayID, archivedFile.Name)
		}
		if written != archivedFile.Size || !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), archivedFile.SHA256) {
			return invalidArchive("题目 %s 的测试文件校验失败：%s", problem.DisplayID, archivedFile.Name)
		}
	}
	return nil
}
