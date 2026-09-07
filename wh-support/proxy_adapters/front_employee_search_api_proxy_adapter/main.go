package main

import "gitlab.wildberries.ru/wbwh/wh-core/gocore_proxy_adapter.git/service"

func main() {
	proxy := service.NewProxyAdapter()
	proxy.Run()
}
