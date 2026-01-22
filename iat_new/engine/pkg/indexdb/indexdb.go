package indexdb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	openOnce sync.Once
	openDB   *leveldb.DB
	openErr  error
	openPath string
)

type ProjectMeta struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func getAppDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(homeDir, ".iat")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return appDir, nil
}

func DefaultDBPath() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("IAT_INDEXDB_DIR")); dir != "" {
		return filepath.Clean(dir), nil
	}
	appDir, err := getAppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDir, "indexdb"), nil
}

func OpenDefault() (*leveldb.DB, error) {
	openOnce.Do(func() {
		dbPath, err := DefaultDBPath()
		if err != nil {
			openErr = err
			return
		}
		openPath = dbPath
		openDB, openErr = leveldb.OpenFile(dbPath, nil)
	})
	return openDB, openErr
}

func OpenedPath() string {
	return openPath
}

func CloseDefault() error {
	if openDB == nil {
		return nil
	}
	err := openDB.Close()
	openDB = nil
	return err
}

func IndexProject(id uint, name string, path string) error {
	db, err := OpenDefault()
	if err != nil {
		return err
	}

	projectKey := []byte(fmt.Sprintf("proj:%d", id))
	prevTokensKey := []byte(fmt.Sprintf("ptok:%d", id))

	var prevTokens []string
	if raw, gerr := db.Get(prevTokensKey, nil); gerr == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &prevTokens)
	}

	nextTokens := tokenizeForProject(name, path)
	prevSet := make(map[string]struct{}, len(prevTokens))
	nextSet := make(map[string]struct{}, len(nextTokens))
	for _, t := range prevTokens {
		prevSet[t] = struct{}{}
	}
	for _, t := range nextTokens {
		nextSet[t] = struct{}{}
	}

	batch := new(leveldb.Batch)

	meta, _ := json.Marshal(ProjectMeta{
		ID:   id,
		Name: name,
		Path: path,
	})
	batch.Put(projectKey, meta)

	nextTokensRaw, _ := json.Marshal(nextTokens)
	batch.Put(prevTokensKey, nextTokensRaw)

	for tok := range prevSet {
		if _, still := nextSet[tok]; still {
			continue
		}
		postingKey := []byte("tok:" + tok)
		ids := getPostingIDs(db, postingKey)
		ids = removeUint(ids, id)
		if len(ids) == 0 {
			batch.Delete(postingKey)
		} else {
			raw, _ := json.Marshal(ids)
			batch.Put(postingKey, raw)
		}
	}

	for tok := range nextSet {
		postingKey := []byte("tok:" + tok)
		ids := getPostingIDs(db, postingKey)
		if !containsUint(ids, id) {
			ids = append(ids, id)
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			raw, _ := json.Marshal(ids)
			batch.Put(postingKey, raw)
		}
	}

	return db.Write(batch, nil)
}

func SearchProjectIDs(query string) ([]uint, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, nil
	}

	db, err := OpenDefault()
	if err != nil {
		return nil, err
	}

	projectIDSet := make(map[uint]struct{})
	
	// Prefix search for tokens
	prefix := []byte("tok:" + q)
	iter := db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	for iter.Next() {
		var ids []uint
		if err := json.Unmarshal(iter.Value(), &ids); err == nil {
			for _, id := range ids {
				projectIDSet[id] = struct{}{}
			}
		}
	}
	
	if err := iter.Error(); err != nil {
		return nil, err
	}

	res := make([]uint, 0, len(projectIDSet))
	for id := range projectIDSet {
		res = append(res, id)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res, nil
}

func getPostingIDs(db *leveldb.DB, key []byte) []uint {
	raw, err := db.Get(key, nil)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var ids []uint
	if jerr := json.Unmarshal(raw, &ids); jerr != nil {
		return nil
	}
	return ids
}

func tokenizeForProject(name string, path string) []string {
	var tokens []string
	for _, s := range []string{name, path} {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		tokens = append(tokens, tokenizeQuery(s)...)
	}
	tokens = uniqueStrings(tokens)
	sort.Strings(tokens)
	return tokens
}

func tokenizeQuery(input string) []string {
	in := strings.ToLower(strings.TrimSpace(input))
	if in == "" {
		return nil
	}

	var tokens []string
	var buf []rune

	flush := func() {
		if len(buf) == 0 {
			return
		}
		tokens = append(tokens, string(buf))
		buf = buf[:0]
	}

	for _, r := range in {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if r <= unicode.MaxASCII {
				buf = append(buf, r)
				continue
			}
			flush()
			tokens = append(tokens, string(r))
			continue
		}
		flush()
	}
	flush()

	if len([]rune(in)) <= 32 {
		compact := strings.ReplaceAll(in, " ", "")
		if compact != "" {
			tokens = append(tokens, compact)
		}
	}

	return uniqueStrings(tokens)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it == "" {
			continue
		}
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}

func intersectUintSorted(a []uint, b []uint) []uint {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	i, j := 0, 0
	out := make([]uint, 0, min(len(a), len(b)))
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			out = append(out, a[i])
			i++
			j++
			continue
		}
		if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return out
}

func containsUint(items []uint, v uint) bool {
	for _, it := range items {
		if it == v {
			return true
		}
	}
	return false
}

func removeUint(items []uint, v uint) []uint {
	if len(items) == 0 {
		return nil
	}
	out := items[:0]
	for _, it := range items {
		if it != v {
			out = append(out, it)
		}
	}
	return out
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
