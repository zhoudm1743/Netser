package tcp

// TCPConnectResponse TCP连接响应
type TCPConnectResponse struct {
	SessionID string `json:"sessionId"` // 会话ID
	Host      string `json:"host"`      // 主机地址
	Port      int    `json:"port"`      // 端口
	IsHex     bool   `json:"isHex"`     // 是否使用十六进制模式
	Status    string `json:"status"`    // 连接状态
}

// TCPListenResponse TCP监听响应
type TCPListenResponse struct {
	SessionID string `json:"sessionId"` // 会话ID
	Port      int    `json:"port"`      // 监听端口
	IsHex     bool   `json:"isHex"`     // 是否使用十六进制模式
	Status    string `json:"status"`    // 监听状态
}

// TCPReceiveResponse TCP接收数据响应
type TCPReceiveResponse struct {
	SessionID string `json:"sessionId"` // 会话ID
	Data      string `json:"data"`      // 接收的数据
	IsHex     bool   `json:"isHex"`     // 是否为十六进制数据
	Timestamp int64  `json:"timestamp"` // 接收时间戳
}

// TCPClientConnectResponse TCP客户端连接响应
type TCPClientConnectResponse struct {
	SessionID  string `json:"sessionId"`  // 会话ID
	ClientAddr string `json:"clientAddr"` // 客户端地址
	IsHex      bool   `json:"isHex"`      // 是否使用十六进制模式
}

// TCPSessionListResponse TCP会话列表响应
type TCPSessionListResponse struct {
	Sessions []TCPSessionInfo `json:"sessions"` // 会话列表
}

// TCPSessionInfo TCP会话信息
type TCPSessionInfo struct {
	SessionID   string `json:"sessionId"`   // 会话ID
	Type        string `json:"type"`        // 会话类型: "client" 或 "server"
	Host        string `json:"host"`        // 主机地址
	Port        int    `json:"port"`        // 端口
	Status      string `json:"status"`      // 状态
	IsHex       bool   `json:"isHex"`       // 是否使用十六进制模式
	ConnectTime int64  `json:"connectTime"` // 连接时间
}
