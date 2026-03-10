package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"pxehub/internal/db"
	"pxehub/internal/dnsmasq"
	httpserver "pxehub/internal/http"
)

func readConf(path string) (map[string]string, error) {
	conf := make(map[string]string)

	// Step 1: load all environment variables first
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		conf[parts[0]] = parts[1]
	}

	// Step 2: optionally load from file if it exists
	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// Only set if not already in env
			if _, exists := conf[key]; !exists {
				conf[key] = val
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Step 3: validate required keys
	required := []string{
		"HTTP_BIND",
		"INTERFACE",
		"DHCP_RANGE_START",
		"DHCP_RANGE_END",
		"DHCP_MASK",
		"DHCP_ROUTER",
		"DNS_SERVERS",
	}

	for _, key := range required {
		if val, ok := conf[key]; !ok || strings.TrimSpace(val) == "" {
			return nil, fmt.Errorf("missing required configuration: %s", key)
		}
	}

	return conf, nil
}

func main() {
	dirs := []string{
		"/opt/pxehub",
		"/opt/pxehub/http",
		"/opt/pxehub/tftp",
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Failed to create %s: %v\n", dir, err)
				continue
			}
			log.Printf("Created %s\n", dir)
		} else if err != nil {
			log.Printf("Error checking %s: %v\n", dir, err)
		} else {
		}
	}

	confPath := os.Getenv("PXEHUB_CONF")
	if confPath == "" {
		confPath = "/opt/pxehub/pxehub.conf"
	}

	conf, err := readConf(confPath)
	if err != nil {
		log.Println("Error reading conf:", err)
		return
	}

	dnsList := []string{}
	if val, ok := conf["DNS_SERVERS"]; ok {
		for _, ip := range strings.Split(val, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				dnsList = append(dnsList, ip)
			}
		}
	}

	database := db.OpenDB("/opt/pxehub/pxehub.db")

	dhcpTftpServer := dnsmasq.DnsmasqServer{
		Iface:       conf["INTERFACE"],
		RangeStart:  conf["DHCP_RANGE_START"],
		RangeEnd:    conf["DHCP_RANGE_END"],
		Mask:        conf["DHCP_MASK"],
		Router:      conf["DHCP_ROUTER"],
		Nameservers: dnsList,
		TFTPDir:     "/opt/pxehub/tftp",
	}

	httpServer := httpserver.HttpServer{
		Address:   conf["HTTP_BIND"],
		Database:  database,
		ExtrasDir: "/opt/pxehub/http",
	}

	if err := dhcpTftpServer.Start(); err != nil {
		fmt.Printf("dnsmasq failed: %v", err)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}

	if err := httpServer.Start(); err != nil {
		fmt.Printf("http failed: %v", err)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Shutting down dnsmasq...")
	if err := dhcpTftpServer.Stop(); err != nil {
		log.Printf("failed to stop dnsmasq: %v", err)
	}
	log.Println("Shutting down http...")
	if err := httpServer.Stop(); err != nil {
		log.Printf("failed to stop http: %v", err)
	}
}
