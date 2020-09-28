package http

import "golang.org/x/net/proxy"

type NewDialler func(proxy interface{}) proxy.Dialer
type ProxyDiallerPool interface {
	GetDialer() proxy.Dialer
	Add(url string)
	SetMode(int)
}

var (
	DefaultProxyDialer NewDialler
	DefaultProxyPool   ProxyDiallerPool
)

func SetProxyGenerater(proxyCreator NewDialler) {
	DefaultProxyDialer = proxyCreator
}
