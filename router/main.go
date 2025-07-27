package router

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zhoudm1743/Netser/dto"
)

func Handle(ctx context.Context, data string) (string, error) {
	request := dto.BaseRequest{}
	err := request.Unmarshal(data)
	if err != nil {
		return "", fmt.Errorf("数据解析失败: %v", err)
	}
	switch request.Name {
	case "get_version":
		return dto.Success("1.0.0"), nil
	case "minimize":
		runtime.WindowMinimise(ctx)
	case "maximize":
		if runtime.WindowIsMaximised(ctx) {
			runtime.WindowUnmaximise(ctx)
		} else {
			runtime.WindowMaximise(ctx)
		}
	case "close":
		runtime.Quit(ctx)
	}
	return "", nil
}
