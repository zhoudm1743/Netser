package session

// SessionListRequest 会话列表请求
type SessionListRequest struct {
	Type string `json:"type"` // 会话类型: "tcp", "udp", "serial", "all"
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"` // 会话列表
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID   string `json:"sessionId"`   // 会话ID
	Type        string `json:"type"`        // 会话类型: "tcp", "udp", "serial"
	Name        string `json:"name"`        // 会话名称
	Status      string `json:"status"`      // 状态
	Host        string `json:"host"`        // 主机地址(对于TCP/UDP客户端)
	Port        int    `json:"port"`        // 端口(对于TCP/UDP)
	Protocol    string `json:"protocol"`    // 协议类型
	IsHex       bool   `json:"isHex"`       // 是否使用十六进制模式
	ConnectTime int64  `json:"connectTime"` // 连接时间
}

// SessionRemoveRequest 移除会话请求
type SessionRemoveRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
}

// SessionRenameRequest 重命名会话请求
type SessionRenameRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
	Name      string `json:"name"`      // 新名称
}

// SessionHistoryRequest 会话历史记录请求
type SessionHistoryRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
	Limit     int    `json:"limit"`     // 限制返回的记录数
	Offset    int    `json:"offset"`    // 偏移量
}

// MessageRecord 消息记录
type MessageRecord struct {
	Direction  string `json:"direction"`  // 方向: "send" 或 "receive"
	Data       string `json:"data"`       // 数据
	IsHex      bool   `json:"isHex"`      // 是否为十六进制数据
	Timestamp  int64  `json:"timestamp"`  // 时间戳
	ByteLength int    `json:"byteLength"` // 字节长度
}

// SessionHistoryResponse 会话历史记录响应
type SessionHistoryResponse struct {
	SessionID string          `json:"sessionId"` // 会话ID
	Records   []MessageRecord `json:"records"`   // 消息记录
	Total     int             `json:"total"`     // 总记录数
}

// SessionClearHistoryRequest 清除会话历史记录请求
type SessionClearHistoryRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
}
