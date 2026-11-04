package judgenode

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alonespig/zeorcode/backend/internal/judgeartifact"
	"github.com/alonespig/zeorcode/backend/internal/repository"
)

const maxArchiveDownloadBytes = int64(1 << 30)

type ArtifactCache struct {
	root       string
	baseURL    *url.URL
	token      string
	httpClient *http.Client
	locks      sync.Map
}

func NewArtifactCache(root, baseURL, token string) (*ArtifactCache, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("judge node cache directory is required")
	}
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("judge node artifact base URL must be an absolute http(s) URL")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("judge node artifact token is required")
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create judge node cache: %w", err)
	}
	return &ArtifactCache{
		root:    filepath.Clean(root),
		baseURL: parsed,
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if req.URL.Scheme != parsed.Scheme || !strings.EqualFold(req.URL.Host, parsed.Host) {
					return errors.New("artifact download redirect leaves configured origin")
				}
				return nil
			},
		},
	}, nil
}

func (c *ArtifactCache) Ensure(ctx context.Context, want judgeartifact.Descriptor) (string, *repository.TestCaseInfo, error) {
	if want.ProblemID <= 0 || !validVersion(want.Version) || !validDigest(want.Digest) {
		return "", nil, errors.New("invalid problem artifact reference")
	}
	key := strconv.FormatInt(want.ProblemID, 10) + ":" + want.Version + ":" + want.Digest
	lockValue, _ := c.locks.LoadOrStore(key, &sync.Mutex{})
	mu := lockValue.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	destination := filepath.Join(c.root, strconv.FormatInt(want.ProblemID, 10), want.Version, want.Digest)
	if info, err := judgeartifact.VerifyDirectory(ctx, destination, want); err == nil {
		return destination, info, nil
	}
	if err := os.RemoveAll(destination); err != nil {
		return "", nil, fmt.Errorf("remove invalid artifact cache: %w", err)
	}
	return c.download(ctx, destination, want)
}

func (c *ArtifactCache) download(ctx context.Context, destination string, want judgeartifact.Descriptor) (string, *repository.TestCaseInfo, error) {
	u := *c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/api/internal/judge/problems/" + strconv.FormatInt(want.ProblemID, 10) + "/artifact"
	query := u.Query()
	query.Set("version", want.Version)
	query.Set("digest", want.Digest)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("download problem artifact: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		return "", nil, fmt.Errorf("download problem artifact: status %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp(c.root, ".artifact-*.zip")
	if err != nil {
		return "", nil, err
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)
	limited := io.LimitReader(resp.Body, maxArchiveDownloadBytes+1)
	written, copyErr := io.Copy(tempFile, limited)
	closeErr := tempFile.Close()
	if copyErr != nil {
		return "", nil, copyErr
	}
	if closeErr != nil {
		return "", nil, closeErr
	}
	if written > maxArchiveDownloadBytes {
		return "", nil, errors.New("problem artifact exceeds download limit")
	}

	tempDir, err := os.MkdirTemp(c.root, ".artifact-unpack-*")
	if err != nil {
		return "", nil, err
	}
	defer os.RemoveAll(tempDir)
	if err := extractArtifact(tempName, tempDir, want); err != nil {
		return "", nil, err
	}
	info, err := judgeartifact.VerifyDirectory(ctx, tempDir, want)
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o750); err != nil {
		return "", nil, err
	}
	if err := os.Rename(tempDir, destination); err != nil {
		return "", nil, fmt.Errorf("activate problem artifact: %w", err)
	}
	return destination, info, nil
}

func extractArtifact(archivePath, destination string, want judgeartifact.Descriptor) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open problem artifact: %w", err)
	}
	defer zr.Close()
	if len(zr.File) == 0 || len(zr.File) > 2002 {
		return errors.New("invalid artifact file count")
	}

	var manifestFound, infoFound bool
	var totalUncompressed uint64
	for _, entry := range zr.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		name := filepath.ToSlash(entry.Name)
		var targetName string
		switch {
		case name == "manifest.json":
			targetName = "manifest.json"
			manifestFound = true
		case name == "info.json":
			targetName = "info.json"
			infoFound = true
		case strings.HasPrefix(name, "cases/"):
			targetName = strings.TrimPrefix(name, "cases/")
			if targetName == "" || filepath.Base(targetName) != targetName || !strings.HasSuffix(targetName, ".in") {
				return fmt.Errorf("invalid artifact entry %q", name)
			}
		default:
			return fmt.Errorf("unexpected artifact entry %q", name)
		}
		if entry.UncompressedSize64 > uint64(maxArchiveDownloadBytes) {
			return errors.New("artifact entry exceeds size limit")
		}
		totalUncompressed += entry.UncompressedSize64
		if totalUncompressed > uint64(maxArchiveDownloadBytes) {
			return errors.New("artifact exceeds uncompressed size limit")
		}
		if err := extractEntry(entry, filepath.Join(destination, targetName)); err != nil {
			return err
		}
	}
	if !manifestFound || !infoFound {
		return errors.New("artifact manifest or testcase info is missing")
	}
	manifestBytes, err := os.ReadFile(filepath.Join(destination, "manifest.json"))
	if err != nil {
		return err
	}
	var got judgeartifact.Descriptor
	if err := json.Unmarshal(manifestBytes, &got); err != nil {
		return errors.New("invalid artifact manifest")
	}
	if got != want {
		return errors.New("artifact manifest does not match requested version")
	}
	return nil
}

func extractEntry(entry *zip.File, target string) error {
	rc, err := entry.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	f, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(f, io.LimitReader(rc, maxArchiveDownloadBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	if written > maxArchiveDownloadBytes {
		return errors.New("artifact entry exceeds size limit")
	}
	return closeErr
}

func validVersion(version string) bool {
	if version == "" || len(version) > 64 || !filepath.IsLocal(version) || filepath.Base(version) != version {
		return false
	}
	for _, r := range version {
		if (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
