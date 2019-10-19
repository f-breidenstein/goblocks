package modules

import (
	"bytes"
	"fmt"
	"net"
	"text/template"

	"github.com/davidscholberg/go-i3barjson"
	"github.com/safchain/ethtool"
)

// Interface represents the configuration for the network interface block.
type Interface struct {
	BlockConfigBase `yaml:",inline"`
	IfaceName       string `yaml:"interface_name"`
	IfaceFormat     string `yaml:"interface_format"`
	IfaceDownFormat string `yaml:"interface_down_format"`
}

// ifaceInfo contains the status info for the interface being monitored.
// The field names correspond directly to the template fields in
// Interface.IfaceFormat.
type ifaceInfo struct {
	Status        string
	Ipv4Addr      string
	Ipv4Cidr      string
	Ipv4LocalAddr string
	Ipv4LocalCidr string
	Ipv6Addr      string
	Ipv6Cidr      string
	Ipv6LocalAddr string
	Ipv6LocalCidr string
}

// UpdateBlock updates the network interface block.
func (c Interface) UpdateBlock(b *i3barjson.Block) {
	var info ifaceInfo

	ethHandle, err := ethtool.NewEthtool()
	if err != nil {
		panic(err.Error())
	}
	defer ethHandle.Close()

	b.Color = c.Color
	fullTextFmt := fmt.Sprintf("%s%%s", c.Label)

	// set default interface_format for backwards compat
	if c.IfaceFormat == "" {
		c.IfaceFormat = "Up"
	}

	if c.IfaceDownFormat == "" {
		c.IfaceDownFormat = "Down"
	}

	iface, err := net.InterfaceByName(c.IfaceName)
	if err != nil {
		b.Urgent = true
		b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
		return
	}

	linkState, _ := ethHandle.LinkState(c.IfaceName)
	if linkState == 0 {
		b.Urgent = true
		info.Status = "down"
	} else {
		b.Urgent = false
		info.Status = "up"
	}

	addrs, err := iface.Addrs()
	if err != nil {
		b.Urgent = true
		b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
		return
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			b.Urgent = true
			b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
			return
		}

		// Checking for address family
		if ip.To4() != nil {
			if ip.IsLinkLocalUnicast() {
				info.Ipv4LocalAddr = ip.String()
				info.Ipv4LocalCidr = addr.String()
			} else {
				info.Ipv4Addr = ip.String()
				info.Ipv4Cidr = addr.String()
			}
		} else {
			if ip.IsLinkLocalUnicast() {
				info.Ipv6LocalAddr = ip.String()
				info.Ipv6LocalCidr = addr.String()
			} else {
				info.Ipv6Addr = ip.String()
				info.Ipv6Cidr = addr.String()
			}
		}
	}

	t := template.New("iface")
	if info.Status == "up" {
		t, err = t.Parse(c.IfaceFormat)
	} else {
		t, err = t.Parse(c.IfaceDownFormat)
	}
	if err != nil {
		b.Urgent = true
		b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
		return
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, info)
	if err != nil {
		b.Urgent = true
		b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
		return
	}

	b.FullText = fmt.Sprintf(fullTextFmt, buf.String())
}
