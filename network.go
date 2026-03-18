package main

import (
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

func setupHostNet(pid int) {

	cleanupNet()

	la := netlink.NewLinkAttrs()
	la.Name = vethHost
	veth := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  vethCont,
	}
	must(netlink.LinkAdd(veth))

	hostVeth, err := netlink.LinkByName(vethHost)
	if err != nil {
		panic(err)
	}
	//mora privatni IP da ne bi bilo konflikta sa stvarnim mrežama :(
	hostAddr, err := netlink.ParseAddr(hostIP)
	if err != nil {
		panic(err)
	}
	must(netlink.AddrAdd(hostVeth, hostAddr))
	must(netlink.LinkSetUp(hostVeth))

	contVeth, err := netlink.LinkByName(vethCont)
	if err != nil {
		panic(err)
	}
	must(netlink.LinkSetNsPid(contVeth, pid))
}

func setupContainerNet() {
	time.Sleep(2 * time.Second)
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		panic(err)
	}
	must(netlink.LinkSetUp(lo))

	contVeth, err := netlink.LinkByName(vethCont)
	if err != nil {
		panic(err)
	}
	contAddr, err := netlink.ParseAddr(contIP)
	if err != nil {
		panic(err)
	}
	must(netlink.AddrAdd(contVeth, contAddr))
	must(netlink.LinkSetUp(contVeth))

	must(netlink.RouteAdd(&netlink.Route{
		LinkIndex: contVeth.Attrs().Index,
		Gw:        net.ParseIP(getawayIp),
	}))
}

func cleanupNet() {
	link, err := netlink.LinkByName(vethHost)
	if err == nil {
		netlink.LinkDel(link)
	}
}
