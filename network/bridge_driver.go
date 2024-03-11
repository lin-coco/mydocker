package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func NewBridgeNetworkDriver() *BridgeNetworkDriver {
	return &BridgeNetworkDriver{}
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}
func (b *BridgeNetworkDriver) Create(nw *Network) (*Network, error) {
	// 配置Linux Bridge
	if err := b.initBridge(nw); err != nil {
		return nil, fmt.Errorf("b.initBridge err: %v", err)
	}
	return nw, nil
}
func (b *BridgeNetworkDriver) Delete(nw *Network) error {
	link, err := netlink.LinkByName(nw.Name)
	if err != nil {
		return fmt.Errorf("netlink.LinkByName err: %v", err)
	}
	if err = netlink.LinkDel(link); err != nil {
		return fmt.Errorf("netlink.LinkDel err %v", err)
	}
	// 删除iptables 规则
	iptablesCmd := fmt.Sprintf("-t nat -D POSTROUTING -s %s ! -o %s -j MASQUERADE", nw.Cidr.String(), nw.Name)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cmd.Output err: %v, output: %s", err, string(output))
	}
	iptablesCmd = fmt.Sprintf("-D FORWARD -i %s -j ACCEPT", nw.Name)
	cmd = exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("cmd.Output err: %v, output: %s", err, string(output))
	}
	return nil
}
func (b *BridgeNetworkDriver) Connect(nw *Network, endpoint *Endpoint) error {
	// 创建veth peer
	bridgeLink, err := netlink.LinkByName(nw.Name)
	if err != nil {
		return fmt.Errorf("netlink.LinkByName err: %v", err)
	}
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]                 // Linux接口名称限制5字符
	la.MasterIndex = bridgeLink.Attrs().Index // 一端挂载在bridge设备上
	veth := netlink.Veth{LinkAttrs: la, PeerName: "cif-" + endpoint.ID[:5]}
	if err = netlink.LinkAdd(&veth); err != nil {
		return fmt.Errorf("netlink.LinkAdd err: %v", err)
	}
	if err = netlink.LinkSetUp(&veth); err != nil {
		return fmt.Errorf("netlink.LinkSetUp err: %v", err)
	}
	peerLink, err := netlink.LinkByName(veth.PeerName)
	if err != nil {
		return fmt.Errorf("netlink.LinkByName err: %v", err)
	}
	endpoint.Device = peerLink
	return nil
}
func (b *BridgeNetworkDriver) Disconnect(nw *Network, deviceName string) error {
	// 删除这对veth设备
	//link, err := netlink.LinkByName(deviceName)
	//if err != nil {
	//	return fmt.Errorf("netlink.LinkByName err: %v", err)
	//}
	//peerLink, err := netlink.LinkByName(link.(*netlink.Veth).PeerName) // 不在主机网络空间中
	//if err != nil {
	//	return fmt.Errorf("netlink.LinkByName err: %v", err)
	//}
	//if err = netlink.LinkDel(link); err != nil {
	//	return fmt.Errorf("netlink.LinkDel err: %v", err)
	//}
	//if err = netlink.LinkDel(peerLink); err != nil {
	//	return fmt.Errorf("netlink.LinkDel err: %v", err)
	//}
	// 从nw中移除设备，释放ip
	for i, device := range nw.Devices {
		if device.Extra == deviceName {
			nw.Devices = append(nw.Devices[:i], nw.Devices[i+1:]...)
			if err := nw.ReleaseIp(device.Addr.To4()); err != nil {
				return fmt.Errorf("nw.ReleaseIp err: %v", err)
			}
			break
		}
	}
	return nil
}

/*
初始化网桥
*/
func (b *BridgeNetworkDriver) initBridge(nw *Network) error {
	/*
		1. 创建bridge虚拟设备
		ip link add testbridge type bridge
	*/
	bridgeName := nw.Name
	_, err := net.InterfaceByName(bridgeName)
	if err != nil && !strings.Contains(err.Error(), "no such network interface") { // 检查是否有同名的设备
		return fmt.Errorf("net.InterfaceByName err: %v", err)
	}
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	bridge := &netlink.Bridge{LinkAttrs: la}
	if err = netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("netlink.LinkAdd err: %v", err)
	}
	/*
		2. 设置bridge设备的地址
		ip addr add 192.168.101.1/24 dev testbridge

	*/
	link, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("netlink.LinkByName err: %v", err)
	}
	ipNet, err := netlink.ParseIPNet(nw.Cidr.String())
	if err != nil {
		return fmt.Errorf("netlink.ParseIPNet err: %v", err)
	}
	ipNet.IP = net.ParseIP(nw.Gateway.To4().String())
	if err = netlink.AddrAdd(link, &netlink.Addr{IPNet: ipNet}); err != nil {
		return fmt.Errorf("netlink.AddrAdd err: %v", err)
	}
	if err = nw.AllocateSpecificIp(ipNet.IP.To4()); err != nil {
		return fmt.Errorf("nw.AllocateSpecificI err: %v", err)
	}
	/*
		3. 启动bridge设备
	*/
	if err = netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("netlink.LinkSetUp err: %v", err)
	}
	/*
			4. 设置iptables的SNAT规则(MASQUERADE)
		    iptables -t nat -A POSTROUTING -s <gatewayIp> ! -o <bridgeName> -j MASQUERADE
	*/
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", ipNet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cmd.Output err: %v, output: %s", err, string(output))
	}
	/*
		5. 允许桥转发
		iptables -A FORWARD -i testbridge -j ACCEPT
	*/
	iptablesCmd = fmt.Sprintf("-A FORWARD -i %s -j ACCEPT", bridgeName)
	cmd = exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("cmd.Output err: %v, output: %s", err, string(output))
	}
	return nil
}
