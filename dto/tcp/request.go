package tcp

// TCPConnectRequest TCP连接请求
type TCPConnectRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
	Host      string `json:"host"`      // 主机地址
	Port      int    `json:"port"`      // 端口
	IsHex     bool   `json:"isHex"`     // 是否使用十六进制模式
	Timeout   int    `json:"timeout"`   // 超时时间(秒)
}

// TCPListenRequest TCP监听请求
type TCPListenRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
	Port      int    `json:"port"`      // 监听端口
	IsHex     bool   `json:"isHex"`     // 是否使用十六进制模式
}

// TCPSendRequest TCP发送数据请求
type TCPSendRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
	Data      string `json:"data"`      // 发送的数据
	IsHex     bool   `json:"isHex"`     // 是否为十六进制数据
}

// TCPDisconnectRequest TCP断开连接请求
type TCPDisconnectRequest struct {
	SessionID string `json:"sessionId"` // 会话ID
}

// TCPCreateSessionRequest TCP创建会话请求
type TCPCreateSessionRequest struct {
	Name    string `json:"name"`    // 会话名称
	Type    string `json:"type"`    // 会话类型: "tcpServer" 或 "tcpClient"
	Host    string `json:"host"`    // 主机地址(仅客户端)
	Port    int    `json:"port"`    // 端口
	IsHex   bool   `json:"isHex"`   // 是否使用十六进制模式
	Timeout int    `json:"timeout"` // 超时时间(秒)
}
