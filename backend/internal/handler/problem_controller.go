package handler

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"zoj/internal/common/errcode"
	"zoj/internal/dto"
	"zoj/internal/judgeworker"
	"zoj/internal/service"

	"github.com/gin-gonic/gin"
)

// regenTestInfo 在上传/改动测试数据后立即重建 info.json（测试点清单）。
// 成功 → 清单即时就位，首次判题无需再算哈希；失败 → 删掉旧清单，交由判题时懒生成兜底（避免留下过期哈希）。
func regenTestInfo(dir string) {
	if err := judgeworker.GenerateInfo(dir); err != nil {
		_ = os.Remove(filepath.Join(dir, "info.json"))
	}
}

type ProblemController struct {
	problemSrv *service.ProblemService
}

func NewProblemController(srv *service.ProblemService) *ProblemController {
	return &ProblemController{problemSrv: srv}
}

func (p *ProblemController) CreateProblem(c *gin.Context) (any, error) {
	var req dto.CreateProblemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	id, err := p.problemSrv.Create(c.Request.Context(), &req)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": id}, nil
}

// resolveProblemPK 把 URL 里的对外题号(:id)解析成内部主键；找不到返回题目不存在。
func (p *ProblemController) resolveProblemPK(c *gin.Context) (int64, error) {
	displayID := c.Param("id")
	if displayID == "" {
		return 0, errcode.ErrInvalidParams.WithMsg("非法题号")
	}
	return p.problemSrv.ResolveID(c.Request.Context(), displayID)
}

func (p *ProblemController) UpdateProblem(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	var req dto.UpdateProblemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	req.ID = pk
	problemID, err := p.problemSrv.Update(c.Request.Context(), &req)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": problemID}, nil
}

func (p *ProblemController) List(c *gin.Context) (any, error) {
	var req dto.ProblemListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}

	userIDStr, _ := c.Get("userID")
	userID, _ := userIDStr.(int64)
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(int)
	isAdmin := role == 1

	return p.problemSrv.List(c.Request.Context(), &req, &userID, isAdmin)
}

func (p *ProblemController) GetProblemDetail(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(int)
	isAdmin := role == 1
	return p.problemSrv.GetByID(c.Request.Context(), pk, isAdmin)
}

func (p *ProblemController) GetProblemForEdit(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	return p.problemSrv.GetForEdit(c.Request.Context(), pk)
}

func (p *ProblemController) GetTagList(c *gin.Context) (any, error) {
	resp, err := p.problemSrv.GetTagList(c.Request.Context())
	if err != nil {
		return nil, err
	}
	return gin.H{"tags": resp.Tags}, nil
}

func (p *ProblemController) GetAdminTagList(c *gin.Context) (any, error) {
	tags, err := p.problemSrv.GetAdminTagList(c.Request.Context())
	if err != nil {
		return nil, err
	}
	return gin.H{"tags": tags}, nil
}

func (p *ProblemController) CreateTag(c *gin.Context) (any, error) {
	var req dto.SaveTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	tag, err := p.problemSrv.CreateTag(c.Request.Context(), req.Name)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (p *ProblemController) UpdateTag(c *gin.Context) (any, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("非法标签 ID")
	}
	var req dto.SaveTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}
	tag, err := p.problemSrv.UpdateTag(c.Request.Context(), id, req.Name)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (p *ProblemController) DeleteTag(c *gin.Context) (any, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("非法标签 ID")
	}
	if err := p.problemSrv.DeleteTag(c.Request.Context(), id); err != nil {
		return nil, err
	}
	return nil, nil
}

const TestDataDir = "./ojdata/problems/"

// safeName 检查文件名是否安全（不含路径穿越字符）
func safeName(name string) bool {
	return name != "" &&
		!strings.Contains(name, "..") &&
		!strings.Contains(name, "/") &&
		!strings.Contains(name, "\\")
}

type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

const maxTestDataPreviewSize int64 = 2 << 20

type TestDataPreview struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated"`
}

func (p *ProblemController) ListFiles(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	problemIDStr := strconv.FormatInt(pk, 10)
	entries, err := os.ReadDir(TestDataDir + problemIDStr)
	if err != nil {
		return nil, errcode.ErrInternal.Wrap(err)
	}

	files := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || e.Name() == "info.json" {
			continue
		}
		info, _ := e.Info()
		files = append(files, FileInfo{
			Name: e.Name(),
			Size: info.Size(),
		})
	}
	return files, nil
}

func (p *ProblemController) GetFileContent(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}

	name := c.Query("file")
	if !isTestcaseName(name) {
		return nil, errcode.ErrInvalidParams.WithMsg("非法测试点文件名")
	}

	dir := filepath.Join(TestDataDir, strconv.FormatInt(pk, 10))
	content, size, truncated, err := readTestDataPreview(dir, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errcode.ErrInvalidParams.WithMsg("测试点文件不存在")
		}
		return nil, errcode.ErrInternal.Wrap(err)
	}

	return TestDataPreview{
		Name:      name,
		Size:      size,
		Content:   content,
		Truncated: truncated,
	}, nil
}

func readTestDataPreview(dir, name string) (content string, size int64, truncated bool, err error) {
	if !isTestcaseName(name) {
		return "", 0, false, errors.New("invalid testcase filename")
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", 0, false, err
	}
	defer root.Close()

	file, err := root.Open(name)
	if err != nil {
		return "", 0, false, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", 0, false, err
	}
	if !info.Mode().IsRegular() {
		return "", 0, false, errors.New("testcase is not a regular file")
	}

	data, err := io.ReadAll(io.LimitReader(file, maxTestDataPreviewSize+1))
	if err != nil {
		return "", 0, false, err
	}
	if int64(len(data)) > maxTestDataPreviewSize {
		data = data[:maxTestDataPreviewSize]
		truncated = true
	}
	return string(data), info.Size(), truncated, nil
}

func (p *ProblemController) UploadFile(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	idStr := strconv.FormatInt(pk, 10)
	file, err := c.FormFile("file")
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("缺少文件").Wrap(err)
	}

	dstDir := filepath.Join(TestDataDir, idStr)

	// zip 自动解压：只提取 N.in/N.out 合并进该题目录，不把 zip 本身存进去
	if strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		if file.Size > maxZipFileSize {
			return nil, errcode.ErrInvalidParams.WithMsg("压缩包过大（上限 200MB）")
		}
		f, err := file.Open()
		if err != nil {
			return nil, errcode.ErrFileUpload.Wrap(err)
		}
		defer f.Close()
		zr, err := zip.NewReader(f, file.Size)
		if err != nil {
			return nil, errcode.ErrInvalidParams.WithMsg("压缩包损坏或不是有效 zip").Wrap(err)
		}
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return nil, errcode.ErrInternal.Wrap(err)
		}
		extracted, err := extractTestcasesFromZip(zr, dstDir)
		if err != nil {
			return nil, err
		}
		if extracted == 0 {
			return nil, errcode.ErrInvalidParams.WithMsg("压缩包内没有 N.in/N.out 测试点文件")
		}
		regenTestInfo(dstDir) // 数据变了 → 立即重建清单
		pairs, missing := countTestcasePairs(dstDir)
		return gin.H{"count": pairs, "missing": missing}, nil
	}

	// 普通单文件：原样保存
	if !safeName(file.Filename) {
		return nil, errcode.ErrInvalidParams.WithMsg("非法文件名")
	}
	if !isTestcaseName(file.Filename) {
		return nil, errcode.ErrInvalidParams.WithMsg("只能上传 .in 或 .out 测试点文件")
	}
	if file.Size > maxSingleCaseSize {
		return nil, errcode.ErrInvalidParams.WithMsg("单个测试文件过大")
	}
	dst := filepath.Join(dstDir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return nil, errcode.ErrFileUpload.Wrap(err)
	}
	// 测试数据变了 → 立即按新数据重建清单(info.json)，避免哈希过期
	regenTestInfo(dstDir)
	return nil, nil
}

// 测试点压缩包上限（管理员上传，取值从宽，仅防极端/炸弹）
const (
	maxZipFileSize       = 200 << 20 // 压缩包本身 200MB
	maxTestcaseFileCount = 1000      // 最多测试点文件数
	maxUncompressedTotal = 512 << 20 // 解压后总大小 512MB
	maxSingleCaseSize    = 256 << 20 // 单个测试文件 256MB
)

// isTestcaseName accepts any non-empty basename ending in .in or .out.
// Pairing is performed by basename, for example input0.in <-> input0.out.
func isTestcaseName(name string) bool {
	var base string
	switch {
	case strings.HasSuffix(name, ".in"):
		base = strings.TrimSuffix(name, ".in")
	case strings.HasSuffix(name, ".out"):
		base = strings.TrimSuffix(name, ".out")
	default:
		return false
	}
	return base != "" && safeName(name)
}

// extractZipEntry 流式把一个 zip 条目写到 dst（不整包读进内存）
func extractZipEntry(zf *zip.File, dst string) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

// extractTestcasesFromZip 遍历 zip，把所有 N.in/N.out 解压进 targetDir（取 basename，
// 天然扁平化 + 防路径穿越），返回写入的文件数。含单文件/总量/数量上限。
func extractTestcasesFromZip(zr *zip.Reader, targetDir string) (int, error) {
	var total int64
	count := 0
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		name := filepath.Base(zf.Name)
		if !isTestcaseName(name) {
			continue // 只收 N.in/N.out；__MACOSX/.DS_Store/其它命名忽略
		}
		if zf.UncompressedSize64 > maxSingleCaseSize {
			return count, errcode.ErrInvalidParams.WithMsg("单个测试文件过大")
		}
		total += int64(zf.UncompressedSize64)
		if total > maxUncompressedTotal {
			return count, errcode.ErrInvalidParams.WithMsg("解压后总大小超限（512MB）")
		}
		count++
		if count > maxTestcaseFileCount {
			return count, errcode.ErrInvalidParams.WithMsg("测试点文件数过多")
		}
		if err := extractZipEntry(zf, filepath.Join(targetDir, name)); err != nil {
			return count, errcode.ErrInternal.Wrap(err)
		}
	}
	return count, nil
}

// countTestcasePairs 扫描目录里的 N.in/N.out，统计成对测试点与未成对文件（合并上传后据实际目录判定）。
func countTestcasePairs(dir string) (pairs int, missing []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil
	}
	hasIn := map[string]bool{}
	hasOut := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !isTestcaseName(name) {
			continue
		}
		if strings.HasSuffix(name, ".in") {
			hasIn[strings.TrimSuffix(name, ".in")] = true
		} else {
			hasOut[strings.TrimSuffix(name, ".out")] = true
		}
	}
	missing = make([]string, 0)
	for k := range hasIn {
		if hasOut[k] {
			pairs++
		} else {
			missing = append(missing, k+".in")
		}
	}
	for k := range hasOut {
		if !hasIn[k] {
			missing = append(missing, k+".out")
		}
	}
	return pairs, missing
}

// UploadTestcaseZip 上传测试点压缩包(zip)：只提取 N.in/N.out（取 basename，天然扁平化 + 防路径穿越），
// 解压到临时目录 → 配对校验 → 原子替换该题测试数据目录。
func (p *ProblemController) UploadTestcaseZip(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	idStr := strconv.FormatInt(pk, 10)
	fh, err := c.FormFile("file")
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("缺少文件").Wrap(err)
	}
	if !strings.HasSuffix(strings.ToLower(fh.Filename), ".zip") {
		return nil, errcode.ErrInvalidParams.WithMsg("请上传 .zip 压缩包")
	}
	if fh.Size > maxZipFileSize {
		return nil, errcode.ErrInvalidParams.WithMsg("压缩包过大（上限 200MB）")
	}

	f, err := fh.Open()
	if err != nil {
		return nil, errcode.ErrFileUpload.Wrap(err)
	}
	defer f.Close()

	zr, err := zip.NewReader(f, fh.Size)
	if err != nil {
		return nil, errcode.ErrInvalidParams.WithMsg("压缩包损坏或不是有效 zip").Wrap(err)
	}

	// 解压到同父目录下的临时目录，保证之后 rename 是同盘原子操作
	tmpDir := filepath.Join(TestDataDir, idStr+".upload-tmp")
	_ = os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, errcode.ErrInternal.Wrap(err)
	}
	defer os.RemoveAll(tmpDir) // 成功时已 rename 走，这里是清残留

	if _, err := extractTestcasesFromZip(zr, tmpDir); err != nil {
		return nil, err
	}
	pairs, missing := countTestcasePairs(tmpDir)
	if pairs == 0 {
		return nil, errcode.ErrInvalidParams.WithMsg("压缩包内没有成对的 N.in/N.out 测试点")
	}

	// 原子替换：旧目录 → .old，tmp → 正式，删 .old
	dstDir := filepath.Join(TestDataDir, idStr)
	oldDir := filepath.Join(TestDataDir, idStr+".old")
	_ = os.RemoveAll(oldDir)
	if _, statErr := os.Stat(dstDir); statErr == nil {
		if err := os.Rename(dstDir, oldDir); err != nil {
			return nil, errcode.ErrInternal.Wrap(err)
		}
	}
	if err := os.Rename(tmpDir, dstDir); err != nil {
		_ = os.Rename(oldDir, dstDir) // 回滚
		return nil, errcode.ErrInternal.Wrap(err)
	}
	_ = os.RemoveAll(oldDir)

	// 数据已就位 → 立即生成清单(info.json)，首次判题无需再算哈希
	regenTestInfo(dstDir)
	return gin.H{"count": pairs, "missing": missing}, nil
}

// DownloadFiles 直接往 Writer 写 zip 二进制，无法套 Wrap，保留 gin.Context 风格
func (p *ProblemController) DownloadFiles(c *gin.Context) {
	displayID := c.Param("id")
	if displayID == "" {
		c.String(400, "invalid problem id")
		return
	}
	pk, err := p.problemSrv.ResolveID(c.Request.Context(), displayID)
	if err != nil {
		c.String(404, "problem not found")
		return
	}
	id := strconv.FormatInt(pk, 10)

	files := c.Query("files")
	if files == "" {
		c.String(400, "no files")
		return
	}

	fileList := strings.Split(files, ",")

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=testdata.zip")

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	for _, name := range fileList {
		if !safeName(name) {
			continue
		}
		path := filepath.Join(TestDataDir+id, name)
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		w, _ := zipWriter.Create(name)
		_, _ = io.Copy(w, f)
		f.Close()
	}
}

type DeleteReq struct {
	Files []string `json:"files"`
}

func (p *ProblemController) DeleteFiles(c *gin.Context) (any, error) {
	pk, err := p.resolveProblemPK(c)
	if err != nil {
		return nil, err
	}
	id := strconv.FormatInt(pk, 10)
	var req DeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.ErrInvalidParams.Wrap(err)
	}

	var failed []string
	for _, name := range req.Files {
		if !safeName(name) {
			failed = append(failed, name)
			continue
		}
		path := filepath.Join(TestDataDir+id, name)
		if err := os.Remove(path); err != nil {
			failed = append(failed, name)
		}
	}
	// 测试数据变了 → 立即按当前目录重建清单
	regenTestInfo(filepath.Join(TestDataDir, id))
	if len(failed) > 0 {
		return gin.H{"failed": failed}, nil
	}
	return nil, nil
}
