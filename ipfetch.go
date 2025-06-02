package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	ColorReset = "\033[0m"
	ColorBlue  = "\033[38;5;33m"
	ColorWhite = "\033[38;5;15m"
)

//go:embed ascii/*
var asciiFiles embed.FS

type IPInfo struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionName  string  `json:"region_name"`
	CityName    string  `json:"city_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	TimeZone    string  `json:"time_zone"`
	AS          string  `json:"as"`
}

func GetIPInfo(ip string) (*IPInfo, error) {
	url := fmt.Sprintf("https://api.ip2location.io/?ip=%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return &info, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ipfetch <IP address or domain>")
		os.Exit(1)
	}

	input := os.Args[1]

	ip := input
	if net.ParseIP(input) == nil {
		ips, err := net.LookupIP(input)
		if err != nil || len(ips) == 0 {
			fmt.Fprintf(os.Stderr, "Error resolving domain: %v\n", err)
			os.Exit(1)
		}
		ip = ips[0].String()
		fmt.Printf("%s[Resolved]%s %s -> %s\n", ColorBlue, ColorWhite, input, ip)
	}

	info, err := GetIPInfo(ip)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	asciiData, err := asciiFiles.ReadFile(fmt.Sprintf("ascii/%s.txt", strings.ToLower(info.CountryCode)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading ASCII art: %v\n", err)
		os.Exit(1)
	}
	asciiLines := strings.Split(string(asciiData), "\n")

	infoLines := []string{
		fmt.Sprintf("%s	IP:%s          %s", ColorBlue, ColorWhite, ip),
		fmt.Sprintf("%s	Country:%s     %s", ColorBlue, ColorWhite, info.CountryName),
		fmt.Sprintf("%s	Region:%s      %s", ColorBlue, ColorWhite, info.RegionName),
		fmt.Sprintf("%s	City:%s        %s", ColorBlue, ColorWhite, info.CityName),
		fmt.Sprintf("%s	Latitude:%s    %.4f", ColorBlue, ColorWhite, info.Latitude),
		fmt.Sprintf("%s	Longitude:%s   %.4f", ColorBlue, ColorWhite, info.Longitude),
		fmt.Sprintf("%s	Time Zone:%s   UTC%s", ColorBlue, ColorWhite, info.TimeZone),
		fmt.Sprintf("%s	ASN:%s         %s", ColorBlue, ColorWhite, info.AS),
	}

	maxLines := len(asciiLines)
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}

	for i := 0; i < maxLines; i++ {
		ascii := ""
		if i < len(asciiLines) {
			ascii = asciiLines[i]
		}
		info := ""
		if i < len(infoLines) {
			info = infoLines[i]
		}
		fmt.Printf("%s%-40s%s\n", ColorBlue, ascii, ColorReset+info)
	}
}
