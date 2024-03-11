package network

var Drivers = make(map[string]Driver)

/*
初始化加载Drivers
*/
func init() {
	bridgeDriver := NewBridgeNetworkDriver()
	Drivers[bridgeDriver.Name()] = bridgeDriver
}

/*
Driver 网络驱动，不同的驱动对网络的创建、连接、销毁不同
*/
type Driver interface {
	// Name 驱动名
	Name() string
	// Create 创建网络
	Create(network *Network) (*Network, error)
	// Delete 删除网络
	Delete(network *Network) error
	// Connect 连接容器网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error
	// Disconnect 从网络上移除容器网络端点
	Disconnect(network *Network, deviceName string) error
}
