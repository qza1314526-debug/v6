package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Config struct {
	IPv6Address string
	IPv4Count   int
	IPv4List    []string
	Ports       []int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: v6 <start|stop|restart|status>")
		os.Exit(1)
	}

	cmd := os.Args[1]

	config := loadConfig()

	switch cmd {
	case "start":
		startService(config)
	case "stop":
		stopService()
	case "restart":
		stopService()
		startService(config)
	case "status":
		statusService()
	default:
		fmt.Println("Invalid command")
		os.Exit(1)
	}
}

func loadConfig() Config {
	config := Config{
		IPv6Address: os.Getenv("IPV6_ADDRESS"),
	}

	ipCountStr := os.Getenv("IPV4_COUNT")
	if ipCountStr == "" {
		ipCountStr = "1"
	}
	config.IPv4Count, _ = strconv.Atoi(ipCountStr)

	for i := 1; i <= config.IPv4Count; i++ {
		ip := os.Getenv(fmt.Sprintf("IPV4_IP%d", i))
		if ip == "" && i == 1 {
			ip = os.Getenv("IPV4_IP") // 兼容旧版本
		}
		if ip != "" {
			config.IPv4List = append(config.IPv4List, ip)
			config.Ports = append(config.Ports, 100+i) // 默认端口101,102,103...
		}
	}

	return config
}

func startService(config Config) {
	if config.IPv6Address == "" || len(config.IPv4List) == 0 {
		fmt.Println("Error: IPV6_ADDRESS and at least one IPv4 address must be set")
		os.Exit(1)
	}

	for i, ipv4 := range config.IPv4List {
		port := config.Ports[i]
		cmd := exec.Command("ip6tables", "-t", "nat", "-A", "PREROUTING", "-d", config.IPv6Address, "-j", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", ipv4, port))
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error setting up IPv6 to IPv4 %s mapping: %v\n", ipv4, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully mapped IPv6 %s to IPv4 %s on port %d\n", config.IPv6Address, ipv4, port)
	}
}

func stopService() {
	cmd := exec.Command("ip6tables", "-t", "nat", "-F")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error clearing IPv6 NAT rules: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully cleared all IPv6 NAT rules")
}

func statusService() {
	cmd := exec.Command("ip6tables", "-t", "nat", "-L")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error checking IPv6 NAT rules: %v\n", err)
		os.Exit(1)
	}
}
