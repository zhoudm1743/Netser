package serial

// 串口相关事件定义
const (
	EventConnected    = "serial:connected"     // 串口连接事件
	EventDisconnected = "serial:disconnected"  // 串口断开连接事件
	EventDataReceived = "serial:data_received" // 接收数据事件
	EventDataSent     = "serial:data_sent"     // 发送数据事件
	EventError        = "serial:error"         // 错误事件
)

// 串口数据位定义
const (
	DataBits5 = 5
	DataBits6 = 6
	DataBits7 = 7
	DataBits8 = 8
)

// 串口停止位定义
const (
	StopBits1  = 1
	StopBits15 = 15
	StopBits2  = 2
)

// 串口校验位定义
const (
	ParityNone  = 0 // 无校验
	ParityOdd   = 1 // 奇校验
	ParityEven  = 2 // 偶校验
	ParityMark  = 3 // 标记校验
	ParitySpace = 4 // 空格校验
)
