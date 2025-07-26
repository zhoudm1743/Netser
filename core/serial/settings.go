package serial

// Settings 串口配置结构体
type Settings struct {
	PortName     string `json:"portName"`     // 端口名称，如 "COM1"、"/dev/ttyS0" 等
	BaudRate     int    `json:"baudRate"`     // 波特率，如 9600、115200 等
	DataBits     int    `json:"dataBits"`     // 数据位，5、6、7、8
	StopBits     int    `json:"stopBits"`     // 停止位，1、1.5、2
	Parity       int    `json:"parity"`       // 校验位，0:None, 1:Odd, 2:Even, 3:Mark, 4:Space
	HexMode      bool   `json:"hexMode"`      // 是否使用16进制模式
	ReadTimeout  int    `json:"readTimeout"`  // 读取超时时间(毫秒)
	WriteTimeout int    `json:"writeTimeout"` // 写入超时时间(毫秒)
	FlowControl  int    `json:"flowControl"`  // 流控制，0:None, 1:Hardware, 2:Software
	BufferSize   int    `json:"bufferSize"`   // 缓冲区大小
	ReadInterval int    `json:"readInterval"` // 读取间隔(毫秒)
}

// NewDefaultSettings 创建默认串口设置
func NewDefaultSettings() *Settings {
	return &Settings{
		PortName:     "COM1",
		BaudRate:     9600,
		DataBits:     DataBits8,
		StopBits:     StopBits1,
		Parity:       ParityNone,
		HexMode:      false,
		ReadTimeout:  1000,
		WriteTimeout: 1000,
		FlowControl:  0,
		BufferSize:   4096,
		ReadInterval: 10,
	}
}

// Clone 克隆设置
func (s *Settings) Clone() *Settings {
	return &Settings{
		PortName:     s.PortName,
		BaudRate:     s.BaudRate,
		DataBits:     s.DataBits,
		StopBits:     s.StopBits,
		Parity:       s.Parity,
		HexMode:      s.HexMode,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		FlowControl:  s.FlowControl,
		BufferSize:   s.BufferSize,
		ReadInterval: s.ReadInterval,
	}
}
