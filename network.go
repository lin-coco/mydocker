package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"mydocker/container"
	"mydocker/network"
)

// 每次运行前都设置一下ipv4转发
func init() {
	// 执行sysctl命令
	cmd := exec.Command("sysctl", "net.ipv4.ip_forward = 1")
	_, err := cmd.CombinedOutput()
	// 检查错误
	if err != nil {
		log.Warn("error setting sysctl rule:", err)
	}
}

/*
CreateNetwork 创建网络
./mydocker network create --driver bridge --subnet 192.168.101.0/24 --gateway 192.168.101.1 testbridge
*/
func CreateNetwork(driverName, subnet, gateway, networkName string) error {
	// 加载网络配置
	nw := &network.Network{Name: networkName}
	if b, err := nw.Load(); err != nil {
		return fmt.Errorf("nw.Load err: %v", err)
	} else if b {
		return fmt.Errorf("network name exist:%s", networkName)
	}
	_, cidr, err := net.ParseCIDR(subnet) // cidr包含的是初始ip
	if err != nil {
		return fmt.Errorf("net.ParseCIDR err: %v", err)
	}
	gatewayIp := net.ParseIP(gateway).To4()
	if gatewayIp == nil {
		return fmt.Errorf("gatewayIp:%s err", gateway)
	}
	driver, b := network.Drivers[driverName]
	if !b {
		return fmt.Errorf("driver:%s not exist", driverName)
	}
	nw, err = network.NewNetwork(networkName, cidr, gatewayIp, driverName)
	if err != nil {
		return fmt.Errorf("network.NewNetwork err: %v", err)
	}
	// 网络驱动创建子网
	_, err = driver.Create(nw)
	if err != nil {
		return fmt.Errorf("driver.Create err: %v", err)
	}
	// 保存网络配置
	if err = nw.Dump(); err != nil {
		return fmt.Errorf("nw.Dump err:%v", err)
	}
	return nil
}

/*
Connect 连接网络
./mydocker run -it -p 8080:8080 -net testbridge busybox sh
*/
func Connect(networkName string, cInfo *container.Info) error {
	// 加载网络配置
	nw := &network.Network{Name: networkName}
	if b, err := nw.Load(); err != nil {
		return fmt.Errorf("nw.Load err: %v", err)
	} else if !b {
		return fmt.Errorf("network name not exist:%s", networkName)
	}
	driver, b := network.Drivers[nw.Driver]
	if !b {
		return fmt.Errorf("no such driver: %v", nw.Driver)
	}
	peerVethIp, err := nw.AllocateIp()
	if err != nil {
		return fmt.Errorf("nw.AllocateIp err: %v", err)
	}
	endpoint := network.Endpoint{
		ID:           fmt.Sprintf("%s-%s", cInfo.Id, networkName),
		IPAddress:    peerVethIp,
		Network:      nw,
		PortMappings: cInfo.PortMappings,
	}
	if err = driver.Connect(nw, &endpoint); err != nil { // 调用connect创建veth连接网络
		return fmt.Errorf("driver.Connect err: %v", err)
	}
	device := network.Device{
		Name:         endpoint.Device.Attrs().Name,
		Addr:         peerVethIp,
		PortMappings: cInfo.PortMappings,
		Extra:        endpoint.ID[:5],
	}
	if err = configEndpointIpAddressAndRoute(&endpoint, cInfo); err != nil { // 进入netns配置peerVethip地址和默认路由
		return fmt.Errorf("configEndpointIpAddressAndRoute err: %v", err)
	}
	if err = configPortMappings(device.Addr, cInfo); err != nil { // 配置端口映射
		return fmt.Errorf("configPortMappings err: %v", err)
	}
	nw.Devices = append(nw.Devices, device)
	// 保存网络配置
	if err = nw.Dump(); err != nil {
		return fmt.Errorf("nw.Dump err:%v", err)
	}
	return nil
}

/*
DisConnect 从网络上移除设备
*/
func DisConnect(networkName string, cInfo *container.Info) error {
	// 加载网络配置
	nw := &network.Network{Name: networkName}
	if b, err := nw.Load(); err != nil {
		return fmt.Errorf("nw.Load err: %v", err)
	} else if !b {
		return fmt.Errorf("network name not exist:%s", networkName)
	}
	driver, b := network.Drivers[nw.Driver]
	if !b {
		return fmt.Errorf("no such driver: %v", nw.Driver)
	}
	deviceName := fmt.Sprintf("%s-%s", cInfo.Id, networkName)[:5]
	var ipAddr net.IP
	for _, device := range nw.Devices {
		if device.Extra == deviceName {
			ipAddr = device.Addr.To4()
		}
	}
	if err := driver.Disconnect(nw, deviceName); err != nil {
		return fmt.Errorf("driver.Disconnect err: %v", err)
	}
	// 取消端口映射
	if err := cancelPortMappings(ipAddr, cInfo); err != nil { // 配置端口映射
		return fmt.Errorf("configPortMappings err: %v", err)
	}
	// 保存网络配置
	if err := nw.Dump(); err != nil {
		return fmt.Errorf("nw.Dump err:%v", err)
	}
	return nil
}

/*
ListNetwork
./mydocker network list
*/
func ListNetwork() error {
	// 加载所有网络配置
	networks := network.LoadAllNetworks()
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	if _, err := fmt.Fprintf(w, "Name\tIpRange\tDriver\t\n"); err != nil {
		return fmt.Errorf("fmt.Fprintf err: %v", err)
	}
	for _, nw := range networks {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.Cidr.String(), nw.Driver); err != nil {
			return fmt.Errorf("fmt.Fprintf err: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("w.Flush err: %v", err)
	}
	return nil
}

/*
DeleteNetwork
./mydocker network remove testbridge
*/
func DeleteNetwork(networkName string) error {
	// 加载网络配置
	nw := &network.Network{Name: networkName}
	if b, err := nw.Load(); err != nil {
		return fmt.Errorf("nw.Load err: %v", err)
	} else if !b {
		return fmt.Errorf("network name not exist:%s", networkName)
	}
	if len(nw.Devices) > 0 { // 还有设备，不能删除
		return fmt.Errorf("there are still devices connected to the network and cannot be deleted")
	}
	// 调用网络驱动删除网络的设备和配置
	driver, b := network.Drivers[nw.Driver]
	if !b {
		return fmt.Errorf("no such driver: %s", nw.Driver)
	}
	if err := driver.Delete(nw); err != nil {
		return fmt.Errorf("driver.Delete err: %v", err)
	}
	// 删除网络配置
	if err := nw.Remove(); err != nil {
		return fmt.Errorf("nw.Dump err:%v", err)
	}
	return nil
}

func configEndpointIpAddressAndRoute(endpoint *network.Endpoint, cInfo *container.Info) error {
	peerLink := endpoint.Device
	// 将peerVeth移动到容器netns中
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("os.OpenFile err: %v", err)
	}
	if err = netlink.LinkSetNsFd(peerLink, int(f.Fd())); err != nil {
		return fmt.Errorf("netlink.LinkSetNsFd err: %v", err)
	}
	/*
		进入容器netns
	*/
	outerContainerNetns, err := enterContainerNetns(cInfo.Pid)
	if err != nil {
		return fmt.Errorf("enterContainerNetns err: %v", err)
	}
	defer outerContainerNetns() // 退出容器netns

	// 设置网络地址、设置默认路由、启动peerVeth、lo
	_, ipNet, err := net.ParseCIDR(endpoint.Network.Cidr.String())
	if err != nil {
		return fmt.Errorf("net.ParseCIDR err: %v", err)
	}
	ipNet.IP = endpoint.IPAddress
	if err = netlink.AddrAdd(peerLink, &netlink.Addr{
		IPNet: ipNet,
	}); err != nil {
		return fmt.Errorf("netlink.AddrAdd err: %v", err)
	}
	loLink, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("netlink.LinkByName err: %v", err)
	}
	if err = netlink.LinkSetUp(peerLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp err: %v", err)
	}
	if err = netlink.LinkSetUp(loLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp err: %v", err)
	}
	_, all, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return fmt.Errorf("net.ParseCIDR err: %v", err)
	}
	if err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        endpoint.Network.Gateway,
		Dst:       all,
	}); err != nil {
		return fmt.Errorf("netlink.RouteAdd err: %v", err)
	}
	if err = netlink.LinkSetUp(peerLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp err: %v", err)
	}
	return nil
}

func configPortMappings(ipAddress net.IP, cInfo *container.Info) error {
	// 设置端口映射
	for _, pm := range cInfo.PortMappings {
		hostPort := pm[0]
		containerPort := pm[1]
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", hostPort, ipAddress.To4().String(), containerPort)
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("cmd.Output err: %v", err)
		}
	}
	return nil
}
func cancelPortMappings(ipAddress net.IP, cInfo *container.Info) error {
	// 设置端口映射
	for _, pm := range cInfo.PortMappings {
		hostPort := pm[0]
		containerPort := pm[1]
		iptablesCmd := fmt.Sprintf("-t nat -D PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", hostPort, ipAddress.To4().String(), containerPort)
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("cmd.Output err: %v", err)
		}
	}
	return nil
}

func enterContainerNetns(pid string) (func(), error) {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", pid), os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile err: %v", err)
	}
	nsFd := f.Fd()
	runtime.LockOSThread() // 锁住当前操作系统线程
	originNs, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("netns.Get err: %v", err)
	}
	if err = netns.Set(netns.NsHandle(nsFd)); err != nil {
		return nil, fmt.Errorf("netns.Set err: %v", err)
	}
	return func() {
		if err = netns.Set(originNs); err != nil {
			log.Error("netns.Set err: %v", err)
		}
		runtime.UnlockOSThread()
		_ = originNs.Close()
		_ = f.Close()
	}, nil
}
