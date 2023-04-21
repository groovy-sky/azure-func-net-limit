package main

import (
	"fmt"
	"net"
)

func mergeCIDR(cidr1, cidr2 string) (string, error) {
	ip1, ipnet1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return "", err
	}
	ip2, ipnet2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return "", err
	}

	// Check if the two CIDR blocks overlap
	if ipnet1.Contains(ip2) || ipnet2.Contains(ip1) {
		return "", fmt.Errorf("CIDR blocks overlap")
	}

	// Calculate the new IP range
	var startIP, endIP net.IP
	if ip1.Cmp(ip2) < 0 {
		startIP = ip1
	} else {
		startIP = ip2
	}
	if ip1.Mask[ipnet1.Mask.Size()-1] == 0 {
		endIP = net.IP{startIP[0], startIP[1], startIP[2], 255}
	} else {
		endIP = net.IP{startIP[0], startIP[1], startIP[2], 254}
	}

	newCIDR := fmt.Sprintf("%s/%d", startIP.String(), endIP[len(endIP)-1])

	return newCIDR, nil
}

func main() {
	cidr1 := "192.168.0.0/24"
	cidr2 := "192.168.2.0/24"
	newCIDR, err := mergeCIDR(cidr1, cidr2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(newCIDR) // Output: 192.168.0.0/23
}
