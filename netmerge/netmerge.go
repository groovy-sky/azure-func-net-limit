package netmerge

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

func closestVectors(in *[]IPv4Vector) (IPv4Vector, IPv4Vector, error) {

	var firstIndex, lastIndex int
	if len(*in) < 2 {
		return IPv4Vector{}, IPv4Vector{}, fmt.Errorf("not enough vectors")
	}

	closest1 := IPv4Vector{}
	closest2 := IPv4Vector{}
	closestDist := math.Inf(1)

	for i := 0; i < len(*in)-1; i++ {
		for j := i + 1; j < len(*in); j++ {
			dist := float64(distance((*in)[i], (*in)[j]))
			if dist < closestDist {
				closest1 = (*in)[i]
				closest2 = (*in)[j]
				closestDist = dist
				if i > j {
					firstIndex = j
					lastIndex = i
				} else {
					firstIndex = i
					lastIndex = j
				}

			}
		}
	}

	*in = append((*in)[:lastIndex], (*in)[lastIndex+1:]...)
	*in = append((*in)[:firstIndex], (*in)[firstIndex+1:]...)

	return closest1, closest2, nil
}

func distance(v1, v2 IPv4Vector) uint32 {
	var minIP, maxIP uint32
	if v1.FirstIP > v2.FirstIP {
		minIP = v1.FirstIP - v2.FirstIP
	} else {
		minIP = v2.FirstIP - v1.FirstIP
	}

	if v1.LastIP > v2.LastIP {
		maxIP = v1.LastIP - v2.LastIP
	} else {
		maxIP = v2.LastIP - v1.LastIP
	}
	return uint32(math.Abs(float64(int32(minIP)))) +
		uint32(math.Abs(float64(int32(maxIP))))
}

func mergeIPNets(v1, v2 *IPv4Vector) (out IPv4Vector, err error) {
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

	newMask := 32 - countDifferentBits(minIP, maxIP)
	newIP := binaryToIP(minIP).To4()
	newCIDR := fmt.Sprintf("%s/%d", newIP, newMask)

	out, err = cidrToVector(newCIDR)

	return out, err
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

func binaryToIP(ip uint32) net.IP {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func uint32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func MergeCIDRs(input []string, maxIpNum uint8) (out []string, err error) {
	var vectors []IPv4Vector
	for _, i := range input {
		v, err := cidrToVector(i)
		if err != nil {
			return out, err
		} else {
			vectors = append(vectors, v)
		}
	}

	var newRange IPv4Vector
	v1, v2, err := closestVectors(&vectors)
	if err != nil {
		return out, err
	}
	newRange, err = mergeIPNets(&v1, &v2)
	if err != nil {
		return out, err
	}
	vectors = append(vectors, newRange)

	for _, v := range vectors {
		ip := uint32ToIP(v.FirstIP)
		mask, _ := v.CIDR.Mask.Size()
		out = append(out, fmt.Sprintf("%s/%d", ip, mask))
	}
	return out, err
}
