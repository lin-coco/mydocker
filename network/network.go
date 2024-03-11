package network

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"os"
	path2 "path"
	"path/filepath"

	"github.com/vishvananda/netlink"

	"mydocker/path"
)

// Networks 网络字段key为网络名称
//var Networks = make(map[string]*Network)

/*
Network 网络相当于bridge上分配的网络，可以让多个容器在这个网络内进行通信
*/
type Network struct {
	Name     string    `json:"name,omitempty"`    // 网络名 (为bridge设备的名称)
	Cidr     net.IPNet `json:"cidr,omitempty"`    // 网段 (地址段中地址为初始地址)
	Gateway  net.IP    `json:"gateway"`           // 网关地址
	IpBitmap Bitmap    `json:"ipBitmap"`          // ip位图
	Driver   string    `json:"driver,omitempty"`  // 网络驱动名
	Devices  []Device  `json:"devices,omitempty"` // 连接到这台网络到设备
}

/*
Endpoint 网络端点，用来连接容器与网络
*/
type Endpoint struct {
	ID           string
	Device       netlink.Link
	IPAddress    net.IP
	MacAddress   net.HardwareAddr
	PortMappings [][]string
	Network      *Network
}

type Device struct {
	Name         string     `json:"name,omitempty"`
	Addr         net.IP     `json:"addr,omitempty"`
	PortMappings [][]string `json:"portMappings,omitempty"`
	Extra        string     `json:"extra,omitempty"` // 额外字段，可以是veth peer设备的另一端
}

func NewNetwork(name string, ipNet *net.IPNet, gateway net.IP, driver string) (*Network, error) {
	if ipNet == nil || gateway == nil {
		return nil, fmt.Errorf("subnet or gateway not be empty")
	}
	one, size := ipNet.Mask.Size()
	ipBitmap := newBitmap(1 << (size - one))
	if !ipNet.Contains(gateway) {
		return nil, fmt.Errorf("this IP:%s is not within the subnet range", gateway.To4().String())
	}
	if _, b := Drivers[driver]; !b {
		return nil, fmt.Errorf("driver:%s not present", driver)
	}
	return &Network{
		Name:     name,
		Cidr:     *ipNet,
		Gateway:  gateway.To4(),
		IpBitmap: ipBitmap,
		Driver:   driver,
	}, nil
}

func (n *Network) AllocateIp() (net.IP, error) {
	ip := net.ParseIP(n.Cidr.IP.String()).To4()
	for i := 0; i < n.IpBitmap.len(); i++ {
		if !n.IpBitmap.get(i) {
			incr := i + 1
			ip[0] += byte(incr >> 24)
			ip[1] += byte(incr >> 16)
			ip[2] += byte(incr >> 8)
			ip[3] += byte(incr >> 0)
			if isBroadcastIP(ip, n.Cidr.Mask) {
				return nil, fmt.Errorf("no more assignable ip")
			}
			n.IpBitmap.set(i)
			break
		}
	}
	return ip, nil
}
func (n *Network) AllocateSpecificIp(ip net.IP) error {
	ip = ip.To4()
	if !n.Cidr.Contains(ip) {
		return fmt.Errorf("%s not contains ipaddr: %s", n.Cidr.String(), ip.String())
	}
	originIp := n.Cidr.IP.To4()
	incr := 0
	incr += int(ip[0]-originIp[0]) << 24
	incr += int(ip[1]-originIp[1]) << 16
	incr += int(ip[2]-originIp[2]) << 8
	incr += int(ip[3]-originIp[3]) << 0
	incr -= 1
	n.IpBitmap.set(incr)
	return nil
}
func (n *Network) ReleaseIp(ip net.IP) error {
	if !n.Cidr.Contains(ip) {
		return fmt.Errorf("%s not contains ipaddr: %s", n.Cidr.String(), ip.String())
	}
	originIp := n.Cidr.IP.To4()
	incr := 0
	incr += int(ip[0]-originIp[0]) << 24
	incr += int(ip[1]-originIp[1]) << 16
	incr += int(ip[2]-originIp[2]) << 8
	incr += int(ip[3]-originIp[3]) << 0
	incr -= 1
	n.IpBitmap.clear(incr)
	return nil
}

func init() {
	networkPath := path.NetworkPath()
	// 检查保存的目录是否存在，不存在就创建
	if _, err := os.Stat(networkPath); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(networkPath, 0644); err != nil {
				panic("os.MkdirAll err: " + err.Error())
			}
		} else {
			panic("os.Stat err: " + err.Error())
		}
	}
}

/*
LoadAllNetworks 加载所有Networks
*/
func LoadAllNetworks() map[string]*Network {
	networks := make(map[string]*Network)
	networkPath := path.NetworkPath()
	// 加载所有network配置
	if err := filepath.Walk(networkPath, func(nwFilePath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		var nw Network
		content, err := os.ReadFile(nwFilePath)
		if err != nil {
			return fmt.Errorf("os.ReadFile err: %v", err)
		}
		if err = json.Unmarshal(content, &nw); err != nil {
			return fmt.Errorf("json.Unmarshal err: %v", err)
		}
		// 将网络的配置信息加载到networks字典中
		networks[nw.Name] = &nw
		return nil
	}); err != nil {
		panic("filepath.Walk err: " + err.Error())
	}
	return networks
}

/*
Dump 保存网络信息
*/
func (n *Network) Dump() error {
	networkPath := path.NetworkPath()
	filePath := path2.Join(networkPath, n.Name+".json")
	content, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("json.Marshal err: %v", err)
	}
	if err = os.WriteFile(filePath, content, 0644); err != nil { // 清空只写入不存在则创建
		return fmt.Errorf("os.WriteFile err: %v", err)
	}
	return nil
}

/*
Load 加载网络信息
*/
func (n *Network) Load() (bool, error) {
	networkPath := path.NetworkPath()
	content, err := os.ReadFile(path2.Join(networkPath, n.Name+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("os.ReadFile err: %v", err)
		}
	}
	if err = json.Unmarshal(content, n); err != nil {
		return true, fmt.Errorf("json.Unmarshal err: %v", err)
	}
	return true, nil
}

/*
Remove 删除网络信息
*/
func (n *Network) Remove() error {
	networkPath := path.NetworkPath()
	if err := os.Remove(path2.Join(networkPath, n.Name+".json")); err != nil {
		return fmt.Errorf("os.Remove err: %v", err)
	}
	return nil
}
func isBroadcastIP(ip net.IP, subnetMask net.IPMask) bool {
	// 获取IP的字节表示
	ipBytes := ip.To4()
	if ipBytes == nil {
		return false // 不是IPv4地址
	}

	// 获取子网掩码的字节表示
	maskBytes := subnetMask
	if len(maskBytes) != net.IPv4len {
		return false // 不是有效的IPv4子网掩码
	}

	// 计算广播地址
	broadcast := make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		broadcast[i] = ipBytes[i] | (^maskBytes[i])
	}

	// 判断是否是广播地址
	return ip.Equal(broadcast)
}
