package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

type IPv4Vector struct {
	FirstIP uint32
	LastIP  uint32
	CIDR    net.IPNet
}

type Vector struct {
	Point1 uint32
	Point2 uint32
}

func cidrToVector(cidr string) (vector IPv4Vector, err error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return vector, err
	}

	// Convert IP address to uint32
	ipUint := binary.BigEndian.Uint32(ip.To4())

	// Calculate the last IP address in the CIDR block
	mask := binary.BigEndian.Uint32(ipNet.Mask)
	lastIP := (ipUint & mask) | (mask ^ 0xffffffff)

	vector = IPv4Vector{ipUint, lastIP, *ipNet}
	return vector, err
}

func ipToUint32(ip net.IP) uint32 {
	return binaryToDecimal(ip.To4())
}

func binaryToDecimal(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func closestVectors(vectors []IPv4Vector) (IPv4Vector, IPv4Vector, error) {
	if len(vectors) < 2 {
		return IPv4Vector{}, IPv4Vector{}, fmt.Errorf("not enough vectors")
	}

	closest1 := IPv4Vector{}
	closest2 := IPv4Vector{}
	closestDist := math.Inf(1)

	for i := 0; i < len(vectors)-1; i++ {
		for j := i + 1; j < len(vectors); j++ {
			dist := float64(distance(vectors[i], vectors[j]))
			if dist < closestDist {
				closest1 = vectors[i]
				closest2 = vectors[j]
				closestDist = dist
			}
		}
	}

	return closest1, closest2, nil
}

func distance(v1, v2 IPv4Vector) uint32 {
	return uint32(math.Abs(float64(int32(v1.FirstIP-v2.FirstIP)))) +
		uint32(math.Abs(float64(int32(v1.LastIP-v2.LastIP))))
}

func mergeIPNets(v1, v2 *IPv4Vector) (uint32, uint32, error) {
	var minIP, maxIP uint32
	if v1.FirstIP > v2.FirstIP {
		minIP = v2.FirstIP
	} else {
		minIP = v1.FirstIP
	}

	if v1.LastIP > v2.LastIP {
		maxIP = v1.LastIP
	} else {
		maxIP = v2.LastIP
	}
	// Create CIDR

	fmt.Println(minIP, "/", 32-countDifferentBits(minIP, maxIP))
	return minIP, maxIP, nil
}

func splitUint32ToBinaryParts(input uint32) []string {
	binaryValue := fmt.Sprintf("%032b", input)
	binaryParts := make([]string, 4)
	for i := 0; i < 4; i++ {
		start := i * 8
		end := (i + 1) * 8
		binaryParts[i] = binaryValue[start:end]
	}
	return binaryParts
}

func countDifferentBits(num1, num2 uint32) int {
	// Convert the numbers to binary strings
	bin1 := fmt.Sprintf("%032b", num1)
	bin2 := fmt.Sprintf("%032b", num2)

	// Compare the binary strings bit by bit
	for i := 0; i < 32; i++ {
		if bin1[i] != bin2[i] {
			// Count the number of bits left till the end of the input
			return 32 - i
		}
	}

	// If all bits match, return 0
	return 0
}

func binaryIP(ip net.IP) uint32 {
	return (uint32(ip[0]) << 24) | (uint32(ip[1]) << 16) | (uint32(ip[2]) << 8) | uint32(ip[3])
}

func binaryToIP(ip uint32) net.IP {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func main() {
	var vectors []IPv4Vector
	for _, i := range []string{"192.168.0.0/27", "192.168.0.64/26"} {
		v, err := cidrToVector(i)
		if err != nil {
			fmt.Println(err)
		} else {
			vectors = append(vectors, v)
		}
	}
	v1, v2, _ := closestVectors(vectors)
	fmt.Println(v1, v2)
	x1, x2, _ := mergeIPNets(&v1, &v2)
	fmt.Println(splitUint32ToBinaryParts(x1))
	fmt.Println(splitUint32ToBinaryParts(x2))
	fmt.Println(countDifferentBits(x1, x2))
}
