package event

// 系统事件常量
const (
	// 应用程序生命周期事件
	EventAppStartup  = "app:startup"
	EventAppShutdown = "app:shutdown"

	// 用户交互事件
	EventUserLogin  = "user:login"
	EventUserLogout = "user:logout"

	// 数据相关事件
	EventDataLoaded  = "data:loaded"
	EventDataSaved   = "data:saved"
	EventDataUpdated = "data:updated"

	// 界面事件
	EventUIChanged = "ui:changed"
	EventUIRefresh = "ui:refresh"
)

// Event 事件数据结构
type Event struct {
	Name   string      `json:"name"`   // 事件名称
	Data   interface{} `json:"data"`   // 事件数据
	Source string      `json:"source"` // 事件来源
}

// NewEvent 创建新事件
func NewEvent(name string, data interface{}, source string) *Event {
	return &Event{
		Name:   name,
		Data:   data,
		Source: source,
	}
}
