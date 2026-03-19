package proto

import "encoding/json"

type MessageType string

const (
	// Agent -> Server
	MsgRegister    MessageType = "register"
	MsgHeartbeat   MessageType = "heartbeat"
	MsgShellOutput MessageType = "shell_output"
	MsgShellExit   MessageType = "shell_exit"
	MsgFileList    MessageType = "file_list"
	MsgFileChunk   MessageType = "file_chunk"
	MsgFileAck     MessageType = "file_ack"
	MsgMetrics     MessageType = "metrics"

	// Server -> Agent
	MsgShellStart      MessageType = "shell_start"
	MsgShellInput      MessageType = "shell_input"
	MsgShellResize     MessageType = "shell_resize"
	MsgShellStop       MessageType = "shell_stop"
	MsgFileListReq     MessageType = "file_list_req"
	MsgFileDownloadReq MessageType = "file_download_req"
	MsgFileUpload      MessageType = "file_upload"
	MsgFileDelete      MessageType = "file_delete"
)

type Envelope struct {
	Type      MessageType     `json:"type"`
	SessionID string          `json:"session_id,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type RegisterPayload struct {
	EnrollToken string `json:"enroll_token"` // server's global enroll token
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
}

type HeartbeatPayload struct {
	Timestamp int64 `json:"timestamp"`
}

type ShellStartPayload struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

type ShellInputPayload struct {
	Data string `json:"data"`
}

type ShellOutputPayload struct {
	Data string `json:"data"`
}

type ShellResizePayload struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

type ShellExitPayload struct {
	Code int `json:"code"`
}

type MetricsPayload struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemTotal    uint64  `json:"mem_total"`
	MemUsed     uint64  `json:"mem_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	NetBytesSent uint64 `json:"net_bytes_sent"`
	NetBytesRecv uint64 `json:"net_bytes_recv"`
	Timestamp   int64   `json:"timestamp"`
}

type FileListPayload struct {
	Path    string     `json:"path"`
	Entries []FileEntry `json:"entries"`
}

type FileEntry struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	Mode    string `json:"mode"`
}

type FileRequestPayload struct {
	Path string `json:"path"`
}

type FileChunkPayload struct {
	Path  string `json:"path"`
	Index int    `json:"index"`
	Total int    `json:"total"`
	Data  []byte `json:"data"`
}

type FileUploadPayload struct {
	Path string `json:"path"`
	Data []byte `json:"data"`
}

type FileDeletePayload struct {
	Path string `json:"path"`
}

type FileAckPayload struct {
	Error string `json:"error,omitempty"`
}
