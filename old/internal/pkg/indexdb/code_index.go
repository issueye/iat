package indexdb

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type ProjectCodeIndexInfo struct {
	ProjectID   uint   `json:"projectId"`
	IndexedAt   string `json:"indexedAt"`
	Files       int    `json:"files"`
	TotalTokens int    `json:"totalTokens"`
}

func ClearProjectCodeIndex(projectID uint) error {
	db, err := OpenDefault()
	if err != nil {
		return err
	}
	prefixes := []string{
		fmt.Sprintf("tokf:%d:", projectID),
		fmt.Sprintf("doc:%d:", projectID),
		fmt.Sprintf("pcode:%d", projectID),
	}
	for _, p := range prefixes {
		if err := clearPrefix(db, []byte(p)); err != nil {
			return err
		}
	}
	return nil
}

func clearPrefix(db *leveldb.DB, prefix []byte) error {
	iter := db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	batch := new(leveldb.Batch)
	for iter.Next() {
		k := iter.Key()
		keyCopy := make([]byte, len(k))
		copy(keyCopy, k)
		batch.Delete(keyCopy)
		if batch.Len() >= 2000 {
			if err := db.Write(batch, nil); err != nil {
				return err
			}
			batch = new(leveldb.Batch)
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}
	if batch.Len() > 0 {
		return db.Write(batch, nil)
	}
	return nil
}

type CodeFileMeta struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	ModTime  string `json:"modTime"`
	TokenCnt int    `json:"tokenCnt"`
}

func IndexProjectCodeFiles(projectID uint, projectPath string, relFiles []string) (*ProjectCodeIndexInfo, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("project path is empty")
	}
	db, err := OpenDefault()
	if err != nil {
		return nil, err
	}
	if err := ClearProjectCodeIndex(projectID); err != nil {
		return nil, err
	}

	tokenToFiles := make(map[string]map[string]struct{}, 4096)
	docBatch := new(leveldb.Batch)
	filesIndexed := 0
	totalTokens := 0

	for _, rel := range relFiles {
		rel = filepath.ToSlash(strings.TrimSpace(rel))
		if rel == "" {
			continue
		}
		abs := filepath.Join(projectPath, filepath.FromSlash(rel))
		info, statErr := os.Stat(abs)
		if statErr != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		if info.Size() > 2*1024*1024 {
			continue
		}
		if !looksLikeTextFile(abs, info.Size()) {
			continue
		}
		content, rerr := os.ReadFile(abs)
		if rerr != nil {
			continue
		}
		text := string(content)
		tokens := tokenizeCodeText(text)
		if len(tokens) == 0 {
			continue
		}
		filesIndexed++
		totalTokens += len(tokens)

		for _, tok := range tokens {
			m, ok := tokenToFiles[tok]
			if !ok {
				m = make(map[string]struct{}, 8)
				tokenToFiles[tok] = m
			}
			m[rel] = struct{}{}
		}

		meta, _ := json.Marshal(CodeFileMeta{
			Path:     rel,
			Size:     info.Size(),
			ModTime:  info.ModTime().UTC().Format(time.RFC3339),
			TokenCnt: len(tokens),
		})
		docBatch.Put([]byte(fmt.Sprintf("doc:%d:%s", projectID, rel)), meta)
		if docBatch.Len() >= 2000 {
			if err := db.Write(docBatch, nil); err != nil {
				return nil, err
			}
			docBatch = new(leveldb.Batch)
		}
	}

	if docBatch.Len() > 0 {
		if err := db.Write(docBatch, nil); err != nil {
			return nil, err
		}
	}

	postBatch := new(leveldb.Batch)
	for tok, filesSet := range tokenToFiles {
		files := make([]string, 0, len(filesSet))
		for f := range filesSet {
			files = append(files, f)
		}
		sort.Strings(files)
		raw, _ := json.Marshal(files)
		postBatch.Put([]byte(fmt.Sprintf("tokf:%d:%s", projectID, tok)), raw)
		if postBatch.Len() >= 2000 {
			if err := db.Write(postBatch, nil); err != nil {
				return nil, err
			}
			postBatch = new(leveldb.Batch)
		}
	}
	if postBatch.Len() > 0 {
		if err := db.Write(postBatch, nil); err != nil {
			return nil, err
		}
	}

	info, _ := json.Marshal(ProjectCodeIndexInfo{
		ProjectID:   projectID,
		IndexedAt:   time.Now().UTC().Format(time.RFC3339),
		Files:       filesIndexed,
		TotalTokens: totalTokens,
	})
	if err := db.Put([]byte(fmt.Sprintf("pcode:%d", projectID)), info, nil); err != nil {
		return nil, err
	}

	return &ProjectCodeIndexInfo{
		ProjectID:   projectID,
		IndexedAt:   time.Now().UTC().Format(time.RFC3339),
		Files:       filesIndexed,
		TotalTokens: totalTokens,
	}, nil
}

func looksLikeTextFile(path string, size int64) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]
	if bytesContainsZero(buf) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return true
	}
	switch ext {
	case ".go", ".js", ".ts", ".tsx", ".vue", ".json", ".md", ".txt", ".py", ".java", ".kt", ".rs", ".c", ".cc", ".cpp", ".h", ".hpp", ".cs", ".html", ".css", ".scss", ".less", ".yml", ".yaml", ".toml", ".ini", ".xml", ".sql", ".sh", ".bat", ".ps1":
		return true
	default:
		return size < 256*1024
	}
}

func SearchProjectIDsByCodeContent(query string) ([]uint, error) {
	db, err := OpenDefault()
	if err != nil {
		return nil, err
	}
	tokens := tokenizeCodeText(query)
	if len(tokens) == 0 {
		return nil, nil
	}

	projectIDSet := make(map[uint]struct{})
	for _, tok := range tokens {
		// We need to find all projects that have this token in their code
		// Keys are tokf:projectID:token
		// This is tricky because projectID is in the middle.
		// LevelDB doesn't support searching by suffix or middle.
		// However, we can iterate through all pcode: keys to get all project IDs,
		// then check tokf:ID:token for each.
		// Or we can use a global tokf:token:projectID structure.

		// Let's use the pcode: approach for now as there aren't many projects.
		pids, _ := GetAllProjectIDsWithCodeIndex()
		for _, pid := range pids {
			key := []byte(fmt.Sprintf("tokf:%d:%s", pid, tok))
			if raw, gerr := db.Get(key, nil); gerr == nil && len(raw) > 0 {
				projectIDSet[pid] = struct{}{}
			}
		}
	}

	res := make([]uint, 0, len(projectIDSet))
	for id := range projectIDSet {
		res = append(res, id)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, nil
}

func GetAllProjectIDsWithCodeIndex() ([]uint, error) {
	db, err := OpenDefault()
	if err != nil {
		return nil, err
	}
	prefix := []byte("pcode:")
	iter := db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	var ids []uint
	for iter.Next() {
		k := string(iter.Key())
		var id uint
		if n, _ := fmt.Sscanf(k, "pcode:%d", &id); n == 1 {
			ids = append(ids, id)
		}
	}
	return ids, iter.Error()
}

func bytesContainsZero(b []byte) bool {
	for _, x := range b {
		if x == 0 {
			return true
		}
	}
	return false
}

func tokenizeCodeText(input string) []string {
	in := strings.ToLower(input)
	var tokens []string
	var buf []rune
	seen := make(map[string]struct{}, 2048)

	flush := func() {
		if len(buf) == 0 {
			return
		}
		if len(buf) > 64 {
			buf = buf[:64]
		}
		tok := string(buf)
		buf = buf[:0]
		if tok == "" {
			return
		}
		if _, ok := seen[tok]; ok {
			return
		}
		seen[tok] = struct{}{}
		tokens = append(tokens, tok)
	}

	for _, r := range in {
		if r <= unicode.MaxASCII {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
				buf = append(buf, r)
				if len(buf) >= 64 {
					flush()
				}
				continue
			}
			flush()
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			flush()
			tok := string(r)
			if _, ok := seen[tok]; ok {
				continue
			}
			seen[tok] = struct{}{}
			tokens = append(tokens, tok)
			continue
		}
		flush()
	}
	flush()

	if len(tokens) > 3000 {
		tokens = tokens[:3000]
	}
	return tokens
}
