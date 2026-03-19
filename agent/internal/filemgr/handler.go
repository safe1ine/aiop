package filemgr

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileEntry struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	Mode    string `json:"mode"`
}

func ListDir(path string) ([]FileEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, FileEntry{
			Name:    e.Name(),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		})
	}
	return result, nil
}

func ReadFileChunked(path string, chunkSize int) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var chunks [][]byte
	for len(data) > 0 {
		n := chunkSize
		if n > len(data) {
			n = len(data)
		}
		chunks = append(chunks, data[:n])
		data = data[n:]
	}
	if len(chunks) == 0 {
		chunks = [][]byte{{}} // empty file = one empty chunk
	}
	return chunks, nil
}

func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func DeletePath(path string) error {
	return os.Remove(path)
}

// MarshalListResponse returns JSON for a file_list response
func MarshalListResponse(path string, entries []FileEntry) (json.RawMessage, error) {
	type resp struct {
		Path    string      `json:"path"`
		Entries []FileEntry `json:"entries"`
	}
	return json.Marshal(resp{Path: path, Entries: entries})
}
