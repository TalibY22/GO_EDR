// Complete EDR (Endpoint Detection and Response) Agent
// Requires admin/root privileges to run
// Monitors files, processes, network, registry, PowerShell, and DNS

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"

	//
	//"internal/goos"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	
	"io"
	"log"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/process"
)

// Log represen
type Log struct {
	AgentID        string      `json:"agent_id"`
	Timestamp      string      `json:"timestamp"`
	Event          string      `json:"event"`
	Details        string      `json:"details"`
	Severity       string      `json:"severity"`
	AdditionalData interface{} `json:"additional_data,omitempty"`
}

type Output struct {
	AgentID       string `json:"agent_id"`
	Given_command string `json:"given_command"`
	Output        string `json:"output"`
}


type NetworkConnection struct {
	SourceIP    string `json:"source_ip"`
	DestIP      string `json:"dest_ip"`
	SourcePort  int    `json:"source_port"`
	DestPort    int    `json:"dest_port"`
	Protocol    string `json:"protocol"`
	ProcessID   int32  `json:"process_id"`
	ProcessName string `json:"process_name"`
}

// DNSQuery represents a DNS query
type DNSQuery struct {
	Domain      string    `json:"domain"`
	QueryType   string    `json:"query_type"`
	Timestamp   time.Time `json:"timestamp"`
	ProcessID   int32     `json:"process_id"`
	ProcessName string    `json:"process_name"`
}

// PowerShellCommand represents a PowerShell command execution
type PowerShellCommand struct {
	Script      string    `json:"script"`
	CommandLine string    `json:"command_line"`
	User        string    `json:"user"`
	Timestamp   time.Time `json:"timestamp"`
	ProcessID   int32     `json:"process_id"`
}

// SuspiciousProcess represents a potentially malicious process
type SuspiciousProcess struct {
	PID         int32    `json:"pid"`
	Name        string   `json:"name"`
	CommandLine string   `json:"command_line"`
	Parent      int32    `json:"parent_pid"`
	Children    []int32  `json:"child_pids"`
	Connections []string `json:"network_connections"`
}

// Terminal COmmands
type TerminalActivity struct {
	Command     string
	Args        []string
	ProcessID   int
	User        string
	Timestamp   time.Time
	IsSudo      bool
	RealCommand string
	RealUser    string // Original user when using sudo
}

type Command struct {
	Command   string `json:"command"`
	Arguments string `json:"arguments"`
}

type Response struct {
	Data []Command `json:"data"`
}

// Program represents the service
type Program struct {
	exit chan struct{}
}

const (
	backendURL = "http://localhost:8080/logs"
	OutputURL  = "http://localhost:8080/output"
	min        = 300 * time.Second
)

type eventCache struct {
	lastEvents map[string]time.Time
}

// Caching to prevent dupplicate logsv
func newEventCache() *eventCache {
	return &eventCache{
		lastEvents: make(map[string]time.Time),
	}
}

func (c *eventCache) shouldLog(eventKey string) bool {
	lastTime, exists := c.lastEvents[eventKey]
	now := time.Now()

	if !exists || now.Sub(lastTime) > min {
		c.lastEvents[eventKey] = now
		return true
	}
	return false
}

var (
	suspiciousProcessNames = []string{
		"mimikatz", "psexec", "powersploit",
		"bloodhound", "cobalt", "metasploit",
	}

	suspiciousCommandPatterns = []string{
		"-enc", "base64", "bypass", "hidden",
		"downloadstring", "invoke-expression",
		"iex", "webclient", "bitstransfer",
	}
)

// Add a global variable to track the current working directory
var currentWorkingDir string

// Initialize the current working directory at startup
func init() {
	var err error
	currentWorkingDir, err = os.Getwd()
	if err != nil {
		log.Printf("Failed to get current directory: %v", err)
		currentWorkingDir = "/"
	}
}

// Service interface implementation
func (p *Program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *Program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}

// PROGRAM ENTRY POINT
func (p *Program) run() {
	agentID := generateAgentID()

	// Start all monitoring functions
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		go monitorFiles(agentID, get_paths(1))
		fmt.Printf(get_paths(1))
		go monitorFiles(agentID, get_paths(2))
		  //go p.monitorProcesses(agentID)
		go p.monitorNetworkConnections(agentID)
		go p.monitorDNS(agentID)
		go p.Getcommand(agentID)
		go p.detecterminal(agentID)
		//go p.monitorBashHistory(agentID)
		go p.monitorUSBDevices(agentID)
		//go p.detectSuspiciousProcesses(agentID)
		go p.monitorUserBehavior(agentID)
	}

	<-p.exit
}

// File Monitoring Functions
func monitorFiles(agentID, path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Failed to initialize file watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	// Watch directory recursively
	err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(subPath)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Failed to watch directory %s: %v\n", path, err)
		return
	}

	var processedFiles = make(map[string]time.Time)
	for {
		select {
		case event := <-watcher.Events:

			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				severity := determineSeverity(event)
				if lastProcessed, exists := processedFiles[event.Name]; exists && time.Since(lastProcessed) < time.Second {
					return // Skip logging if event is too recent
				}
				log := Log{
					AgentID:   agentID,
					Timestamp: time.Now().Format(time.RFC3339),
					Event:     "file_event",
					Details:   fmt.Sprintf("File event: %s on %s", event.Op, event.Name),
					Severity:  severity,
				}

				sendLog(log)
				processedFiles[event.Name] = time.Now()
			}
		case err := <-watcher.Errors:
			fmt.Printf("Error watching files: %v\n", err)
		}
	}
}

// Process Monitoring Functions
func (p *Program) monitorProcesses(agentID string) {
	lastCheck := make(map[int32]time.Time)

	for {
		processes, err := process.Processes()
		if err != nil {
			continue
		}

		for _, proc := range processes {
			createTime, err := proc.CreateTime()
			if err != nil {
				continue
			}

			lastCheckTime, exists := lastCheck[proc.Pid]
			procCreateTime := time.Unix(createTime/1000, 0)

			if !exists || procCreateTime.After(lastCheckTime) {
				name, _ := proc.Name()
				cmdline, _ := proc.Cmdline()
				username, _ := proc.Username()

				log := Log{
					AgentID:   agentID,
					Timestamp: time.Now().Format(time.RFC3339),
					Event:     "process_created",
					Details:   fmt.Sprintf("New process: %s (PID: %d) by user: %s", name, proc.Pid, username),
					AdditionalData: map[string]interface{}{
						"process_name": name,
						"pid":          proc.Pid,
						"command":      cmdline,
						"user":         username,
					},
				}
				sendLog(log)
				lastCheck[proc.Pid] = time.Now()
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// Network Monitoring Functions
func (p *Program) monitorNetworkConnections(agentID string) {
	cache := newEventCache()
	localNetworks := getLocalNetworks()

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Failed to get network interfaces: %v\n", err)
		return
	}

	for _, iface := range interfaces {
		go func(interfaceName string) {
			handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
			if err != nil {
				return
			}
			defer handle.Close()

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				networkLayer := packet.NetworkLayer()
				if networkLayer == nil {
					continue
				}

				srcIP := networkLayer.NetworkFlow().Src().String()
				dstIP := networkLayer.NetworkFlow().Dst().String()

				// Skip localhost and internal network traffic
				if isInternalTraffic(srcIP, dstIP, localNetworks) {
					continue

				}

				// Create a unique key for this connection
				connKey := fmt.Sprintf("%s_%s", srcIP, dstIP)
				if cache.shouldLog(connKey) {
					connection := NetworkConnection{
						SourceIP: srcIP,
						DestIP:   dstIP,
					}

					log := Log{
						AgentID:        agentID,
						Timestamp:      time.Now().Format(time.RFC3339),
						Event:          "external_network_connection",
						Details:        fmt.Sprintf("External connection: %s -> %s", srcIP, dstIP),
						AdditionalData: connection,
					}
					sendLog(log)
				}
			}
		}(iface.Name)
	}
}

// function to get local networks
func getLocalNetworks() []*net.IPNet {
	var networks []*net.IPNet
	interfaces, err := net.Interfaces()
	if err != nil {
		return networks
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				networks = append(networks, ipnet)
			}
		}
	}

	return networks
}

// function to check if traffic is internal
func isInternalTraffic(srcIP, dstIP string, localNetworks []*net.IPNet) bool {
	// Check if it's localhost traffic
	if srcIP == "127.0.0.1" || dstIP == "127.0.0.1" ||
		srcIP == "::1" || dstIP == "::1" {
		return true
	}

	src := net.ParseIP(srcIP)
	dst := net.ParseIP(dstIP)

	// Check if both IPs are in local networks
	for _, network := range localNetworks {
		if network.Contains(src) && network.Contains(dst) {
			return true
		}
	}

	return false
}

// DNS Monitoring Functions
func (p *Program) monitorDNS(agentID string) {
	interfaces, err := pcap.FindAllDevs()
	if err != nil {
		fmt.Printf("Failed to get network interfaces: %v\n", err)
		return
	}

	for _, iface := range interfaces {
		go func(device string) {
			handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
			if err != nil {
				return
			}
			defer handle.Close()

			err = handle.SetBPFFilter("udp and port 53")
			if err != nil {
				return
			}

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				dnsLayer := packet.Layer(layers.LayerTypeDNS)
				if dnsLayer == nil {
					continue
				}

				dns, _ := dnsLayer.(*layers.DNS)
				if !dns.QR {
					for _, question := range dns.Questions {
						query := DNSQuery{
							Domain:    string(question.Name),
							QueryType: question.Type.String(),
							Timestamp: time.Now(),
						}

						if isSuspiciousDomain(string(question.Name)) {
							log := Log{
								AgentID:        agentID,
								Timestamp:      time.Now().Format(time.RFC3339),
								Event:          "suspicious_dns_query",
								Details:        fmt.Sprintf("Suspicious DNS query: %s", string(question.Name)),
								Severity:       "medium",
								AdditionalData: query,
							}
							sendLog(log)
						}
					}
				}
			}
		}(iface.Name)
	}
}

// Helper Functions

func generateAgentID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-agent"
	}
	return fmt.Sprintf("agent-%s", hostname)
}

//Combine this filepaths function

// 1=Downloads 2 = Desktop
func get_paths(pt int) string {

	if pt == 1 {
		home, err := os.UserHomeDir()

		if err != nil {
			return " "
		}

		return filepath.Join(home, "Downloads")
	} else if pt == 2 {

		home, err := os.UserHomeDir()

		if err != nil {
			return " "
		}

		return filepath.Join(home, "Desktop")
	} else {
		return " "
	}

}

func determineSeverity(event fsnotify.Event) string {
	ext := strings.ToLower(filepath.Ext(event.Name))
	suspiciousExts := map[string]bool{
		".exe": true, ".dll": true, ".ps1": true,
		".bat": true, ".cmd": true, ".scr": true,
	}
	if suspiciousExts[ext] && (event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Create == fsnotify.Create) {
		return "high"
	}
	return "low"
}

func isSuspiciousDomain(domain string) bool {
	suspiciousPatterns := []string{
		".xyz", ".top", ".pw", "pastebin.com",
		"raw.githubusercontent.com", ".bit", ".onion",
	}

	domainLower := strings.ToLower(domain)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(domainLower, pattern) {
			return true
		}
	}

	if isRandomLookingDomain(domain) {
		return true
	}

	return false
}

func isRandomLookingDomain(domain string) bool {
	consonantCount := 0
	numberCount := 0
	domain = strings.ToLower(domain)

	for _, char := range domain {
		if char >= '0' && char <= '9' {
			numberCount++
		}
		if strings.ContainsRune("bcdfghjklmnpqrstvwxz", char) {
			consonantCount++
			if consonantCount > 4 {
				return true
			}
		} else {
			consonantCount = 0
		}
	}

	return numberCount > 5
}

// Detect suspicious processes
func (p *Program) detectSuspiciousProcesses(agentID string) {
	for {
		processes, _ := process.Processes()
		for _, proc := range processes {
			name, err := proc.Name()
			if err != nil {
				continue
			}

			cmdline, _ := proc.Cmdline()

			// Check for suspicious process names
			for _, suspiciousName := range suspiciousProcessNames {
				if strings.Contains(strings.ToLower(name), suspiciousName) {
					suspicious := SuspiciousProcess{
						PID:         proc.Pid,
						Name:        name,
						CommandLine: cmdline,
					}

					log := Log{
						AgentID:        agentID,
						Timestamp:      time.Now().Format(time.RFC3339),
						Event:          "suspicious_process",
						Details:        fmt.Sprintf("Suspicious process detected: %s (PID: %d)", name, proc.Pid),
						Severity:       "high",
						AdditionalData: suspicious,
					}
					sendLog(log)
				}
			}

			// Check for suspicious command line patterns
			for _, pattern := range suspiciousCommandPatterns {
				if strings.Contains(strings.ToLower(cmdline), pattern) {
					log := Log{
						AgentID:   agentID,
						Timestamp: time.Now().Format(time.RFC3339),
						Event:     "suspicious_commandline",
						Details:   fmt.Sprintf("Suspicious command line in process %s: %s", name, cmdline),
						Severity:  "medium",
					}
					sendLog(log)
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Monitor is any usb devices are plugged in
func (p *Program) monitorUSBDevices(agentID string) {
	cache := newEventCache()
	lastDevices := make(map[string]bool)

	for {
		var devices []string
		var err error

		if runtime.GOOS == "linux" {
			// Only monitor actual USB storage devices
			devices, err = filepath.Glob("/sys/block/sd*/removable")
		} else if runtime.GOOS == "windows" {
			cmd := exec.Command("wmic", "diskdrive", "where", "InterfaceType='USB'", "get", "DeviceID")
			output, err := cmd.Output()
			if err == nil {
				devices = strings.Split(string(output), "\n")
			}
		}

		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		currentDevices := make(map[string]bool)
		for _, device := range devices {

			mainDevice := strings.Split(device, ":")[0]
			currentDevices[mainDevice] = true

			if !lastDevices[mainDevice] {
				eventKey := fmt.Sprintf("usb_%s", mainDevice)
				if cache.shouldLog(eventKey) {
					log := Log{
						AgentID:   agentID,
						Timestamp: time.Now().Format(time.RFC3339),
						Event:     "usb_storage_connected",
						Details:   fmt.Sprintf("USB storage device connected: %s", mainDevice),
						Severity:  "medium",
					}
					sendLog(log)
				}
			}
		}

		lastDevices = currentDevices
		time.Sleep(5 * time.Second)
	}
}

// NOtify if a user logs in
func (p *Program) monitorUserBehavior(agentID string) {
	lastLoginCheck := time.Now()

	for {
		processes, _ := process.Processes()
		for _, proc := range processes {
			name, err := proc.Name()
			if err != nil {
				continue
			}

			if name == "login" || name == "sshd" {
				createTime, err := proc.CreateTime()
				if err != nil {
					continue
				}

				procCreateTime := time.Unix(createTime/1000, 0)
				if procCreateTime.After(lastLoginCheck) {
					username, _ := proc.Username()
					log := Log{
						AgentID:   agentID,
						Timestamp: time.Now().Format(time.RFC3339),
						Event:     "user_login",
						Details:   fmt.Sprintf("New login detected for user: %s", username),
						Severity:  "medium",
					}
					sendLog(log)
				}
			}
		}

		lastLoginCheck = time.Now()
		time.Sleep(30 * time.Second)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//RUN COMMANDS RETRIEVE FROM SERVER
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Program) Getcommand(agentid string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Make the HTTP request
			resp, err := http.Get("http://localhost:8080/command")
			if err != nil {
				log.Printf("Failed to get command: %v", err)
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read command: %v", err)
				continue
			}

			var response Response
			if err := json.Unmarshal(body, &response); err != nil {
				log.Printf("Failed to unmarshal command: %v", err)
				continue
			}

			if len(response.Data) == 0 {
				continue // No commands to execute
			}

			command := response.Data[0]
			if command.Command == "" {
				log.Printf("Empty command received")
				continue
			}

			// Check for special commands
			if command.Command == "snap" || command.Command == "screenshot" {
				// Handle screenshot command
				screenshotPath, err := takeScreenshot(agentid)
				if err != nil {
					log.Printf("Failed to take screenshot: %v", err)
					continue
				}

				// Send the screenshot
				err = sendScreenshot(agentid, screenshotPath)
				if err != nil {
					log.Printf("Failed to send screenshot: %v", err)
				}

				// Clean up the temporary file
				os.Remove(screenshotPath)
				continue
			}

			// Check for "send" command to send files
			if command.Command == "send" {
				output, err := sendFilesFromCurrentDirectory(agentid)
				if err != nil {
					log.Printf("Failed to send files: %v", err)

					// Send error as output
					out := Output{
						AgentID:       agentid,
						Given_command: command.Command,
						Output:        fmt.Sprintf("Error sending files: %v", err),
					}
					sendOutput(out)
					continue
				}

				// Send success output
				out := Output{
					AgentID:       agentid,
					Given_command: command.Command,
					Output:        output,
				}
				sendOutput(out)
				continue
			}

			// Check for "cd" command to change directory
			if strings.HasPrefix(command.Command, "cd ") {
				// Extract the directory path
				dirPath := strings.TrimPrefix(command.Command, "cd ")
				dirPath = strings.TrimSpace(dirPath)

				// Handle special cases
				if dirPath == "~" {
					// Get home directory
					homeDir, err := os.UserHomeDir()
					if err != nil {
						out := Output{
							AgentID:       agentid,
							Given_command: command.Command,
							Output:        fmt.Sprintf("Error: Failed to get home directory: %v", err),
						}
						sendOutput(out)
						continue
					}
					dirPath = homeDir
				}

				// If relative path, make it absolute based on current directory
				if !filepath.IsAbs(dirPath) {
					dirPath = filepath.Join(currentWorkingDir, dirPath)
				}

				// Check if directory exists
				fileInfo, err := os.Stat(dirPath)
				if err != nil {
					out := Output{
						AgentID:       agentid,
						Given_command: command.Command,
						Output:        fmt.Sprintf("Error: %v", err),
					}
					sendOutput(out)
					continue
				}

				// Check if it's a directory
				if !fileInfo.IsDir() {
					out := Output{
						AgentID:       agentid,
						Given_command: command.Command,
						Output:        fmt.Sprintf("Error: %s is not a directory", dirPath),
					}
					sendOutput(out)
					continue
				}

				// Change the current working directory
				currentWorkingDir = dirPath

				// Send success output
				out := Output{
					AgentID:       agentid,
					Given_command: command.Command,
					Output:        fmt.Sprintf("Changed directory to: %s", currentWorkingDir),
				}
				sendOutput(out)
				continue
			}

			// Check for "pwd" command to show current directory
			if command.Command == "pwd" {
				out := Output{
					AgentID:       agentid,
					Given_command: command.Command,
					Output:        currentWorkingDir,
				}
				sendOutput(out)
				continue
			}

			// Check if this is a sudo command with a password
			if strings.HasPrefix(command.Command, "sudo ") && command.Arguments != "" {
				// Execute sudo command with password
				output, err := executeSudoCommand(command.Command, command.Arguments)
				if err != nil {
					log.Printf("Failed to execute sudo command: %v", err)

					// Send error as output
					out := Output{
						AgentID:       agentid,
						Given_command: command.Command,
						Output:        fmt.Sprintf("Error: %v", err),
					}
					sendOutput(out)
					continue
				}

				// Send the output
				out := Output{
					AgentID:       agentid,
					Given_command: command.Command,
					Output:        output,
				}
				sendOutput(out)
				continue
			}

			// Execute regular command with the correct working directory
			cmd := exec.Command("bash", "-c", command.Command)
			cmd.Dir = currentWorkingDir // Set the working directory for the command
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Failed to execute command: %v", err)

				// Send error as output
				out := Output{
					AgentID:       agentid,
					Given_command: command.Command,
					Output:        fmt.Sprintf("Error: %v\n%s", err, string(output)),
				}
				sendOutput(out)
				continue
			}

			// Prepare and send output
			out := Output{
				AgentID:       agentid,
				Given_command: command.Command,
				Output:        string(output),
			}
			sendOutput(out)

			log.Printf("Command output: %s", string(output))

		case <-p.exit:
			return
		}
	}
}

// Update the sendFilesFromCurrentDirectory function to use the tracked directory
func sendFilesFromCurrentDirectory(agentID string) (string, error) {
	// Use the current working directory
	files, err := ioutil.ReadDir(currentWorkingDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %v", err)
	}

	// Count of successfully sent files
	sentCount := 0
	failedFiles := []string{}

	// Send each file
	for _, file := range files {
		// Skip directories and very large files
		if file.IsDir() || file.Size() > 50*1024*1024 { // Skip files larger than 50MB
			continue
		}

		filePath := filepath.Join(currentWorkingDir, file.Name())
		err := sendFileToServer(agentID, filePath)
		if err != nil {
			failedFiles = append(failedFiles, file.Name())
			log.Printf("Failed to send file %s: %v", file.Name(), err)
		} else {
			sentCount++
		}
	}

	// Prepare result message
	result := fmt.Sprintf("Successfully sent %d files from %s\n", sentCount, currentWorkingDir)
	if len(failedFiles) > 0 {
		result += fmt.Sprintf("Failed to send %d files: %s", len(failedFiles), strings.Join(failedFiles, ", "))
	}

	return result, nil
}

// Update executeSudoCommand to use the current working directory
func executeSudoCommand(command, password string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = currentWorkingDir // Set the working directory

	// Get stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	// Get stdout and stderr pipes
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %v", err)
	}

	// Write the password to stdin
	io.WriteString(stdin, password+"\n")
	stdin.Close()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return stderr.String(), fmt.Errorf("command failed: %v", err)
	}

	// Return the output
	if stderr.Len() > 0 {
		return stdout.String() + "\n" + stderr.String(), nil
	}
	return stdout.String(), nil
}




//Take a screenshot on the computer itself
func takeScreenshot(agentID string) (string, error) {
	// Create a temporary file to store the screenshot
	tmpfile, err := ioutil.TempFile("", "screenshot-*.png")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpfile.Close()

	cmd := exec.Command("import", "-window", "root", tmpfile.Name())
	err = cmd.Run()

	if err != nil {
		// Try alternative screenshot tools if import fails
		cmd = exec.Command("scrot", tmpfile.Name())
		err = cmd.Run()

		if err != nil {
			// Try gnome-screenshot if scrot fails
			cmd = exec.Command("gnome-screenshot", "-f", tmpfile.Name())
			err = cmd.Run()

			if err != nil {
				return "", fmt.Errorf("failed to take screenshot: %v", err)
			}
		}
	}

	return tmpfile.Name(), nil
}



//Function to send a sceenshot back to the C2
func sendScreenshot(agentID, screenshotPath string) error {

	file, err := os.Open(screenshotPath)
	if err != nil {
		return fmt.Errorf("failed to open screenshot file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err = writer.WriteField("agent_id", agentID)
	if err != nil {
		return fmt.Errorf("failed to write agent ID field: %v", err)
	}

	part, err := writer.CreateFormFile("screenshot", filepath.Base(screenshotPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file to form: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/upload", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	return nil
}






///COMAND MONITORING FUNCTIONS THIS SHIT BREAKS ALL THE TIME 

func removeEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// enrichCommandInfo gathers more details about a command
func enrichCommandInfo(activity *TerminalActivity) {
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", activity.ProcessID)
	
	// Check if the process still exists
	if _, err := os.Stat(cmdlinePath); os.IsNotExist(err) {
		fmt.Printf("Process %d no longer exists\n", activity.ProcessID)
		return
	}

	cmdline, err := os.ReadFile(cmdlinePath)
	if err != nil {
		fmt.Printf("Failed to read command line for PID %d: %v\n", activity.ProcessID, err)
		return
	}

	// Split the command line into parts
	parts := strings.Split(string(cmdline), "\x00")
	parts = removeEmpty(parts)

	if len(parts) == 0 {
		return
	}

	// Extract command name from path if needed
	baseName := filepath.Base(parts[0])
	
	if baseName == "sudo" && len(parts) > 1 {
		activity.IsSudo = true
		activity.RealUser = activity.User
		activity.User = "root"
		activity.Command = baseName
		activity.RealCommand = filepath.Base(parts[1])
		if len(parts) > 2 {
			activity.Args = parts[2:]
		} else {
			activity.Args = []string{}
		}
	} else {
		activity.Command = baseName
		if len(parts) > 1 {
			activity.Args = parts[1:]
		} else {
			activity.Args = []string{}
		}
	}

	// Try to get SUDO_USER from environment if available
	environPath := fmt.Sprintf("/proc/%d/environ", activity.ProcessID)
	if _, err := os.Stat(environPath); !os.IsNotExist(err) {
		environ, err := os.ReadFile(environPath)
		if err == nil {
			env := strings.Split(string(environ), "\x00")
			for _, e := range env {
				if strings.HasPrefix(e, "SUDO_USER=") {
					activity.RealUser = strings.TrimPrefix(e, "SUDO_USER=")
					break
				}
			}
		}
	}
}

// monitorProcessEvents uses /proc to monitor for new processes
func (p *Program) monitorProcessEvents(agentID string) {
	// Track processes we've already seen
	seenPIDs := make(map[int]bool)
	cache := newEventCache()

	for {
		// Read all processes from /proc
		procDirs, err := os.ReadDir("/proc")
		if err != nil {
			fmt.Printf("Error reading /proc: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, dir := range procDirs {
			// Skip non-numeric directories (not processes)
			pid, err := strconv.Atoi(dir.Name())
			if err != nil {
				continue
			}

			// Skip PIDs we've already processed
			if _, seen := seenPIDs[pid]; seen {
				continue
			}

			// Mark this PID as seen
			seenPIDs[pid] = true

			// Read process status to get information
			statusFile := fmt.Sprintf("/proc/%d/status", pid)
			
			// Skip if the process already disappeared
			if _, err := os.Stat(statusFile); os.IsNotExist(err) {
				continue
			}
			
			// Read process information
			statusContent, err := os.ReadFile(statusFile)
			if err != nil {
				continue
			}
			
			// Parse status content
			lines := strings.Split(string(statusContent), "\n")
			var processName string
			var processUID string
			
			for _, line := range lines {
				if strings.HasPrefix(line, "Name:") {
					processName = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
				} else if strings.HasPrefix(line, "Uid:") {
					uidParts := strings.Fields(strings.TrimPrefix(line, "Uid:"))
					if len(uidParts) > 0 {
						processUID = uidParts[0]
					}
				}
			}
			
			// Only process shell commands (bash, sh, zsh, etc.)
			isShell := false
			shellCommands := []string{"bash", "sh", "zsh", "dash", "ksh", "fish"}
			for _, shell := range shellCommands {
				if processName == shell {
					isShell = true
					break
				}
			}
			
			if !isShell {
				continue
			}
			
			// Create a terminal activity object
			username := processUID
			
			// Try to convert UID to username
			if u, err := user.LookupId(processUID); err == nil {
				username = u.Username
			}
			
			activity := &TerminalActivity{
				Command:   processName,
				User:      username,
				ProcessID: pid,
				Timestamp: time.Now(),
			}
			
			// Enrich with additional info
			enrichCommandInfo(activity)
			
			// Create a unique key for deduplication
			cmdKey := fmt.Sprintf("%d_%s_%s", pid, activity.User, activity.Command)
			if activity.RealCommand != "" {
				cmdKey += "_" + activity.RealCommand
			}
			
			if cache.shouldLog(cmdKey) {
				commandStr := activity.Command
				if activity.RealCommand != "" {
					commandStr = fmt.Sprintf("%s (%s %s)", commandStr, activity.RealCommand, strings.Join(activity.Args, " "))
				} else if len(activity.Args) > 0 {
					commandStr = fmt.Sprintf("%s %s", commandStr, strings.Join(activity.Args, " "))
				}
				
				fmt.Printf("Detected terminal command: %s (user: %s)\n", commandStr, activity.User)
				
				log := Log{
					AgentID:        agentID,
					Timestamp:      time.Now().Format(time.RFC3339),
					Event:          "terminal_command",
					Details:        fmt.Sprintf("Terminal command executed: %s", commandStr),
					Severity:       "medium",
					AdditionalData: activity,
				}
				sendLog(log)
			}
		}
		
		// Cleanup old PIDs that no longer exist
		for pid := range seenPIDs {
			procPath := fmt.Sprintf("/proc/%d", pid)
			if _, err := os.Stat(procPath); os.IsNotExist(err) {
				delete(seenPIDs, pid)
			}
		}
		
		// Small delay to prevent high CPU usage
		time.Sleep(500 * time.Millisecond)
	}
}

// Custom DirEntry implementation for root directory (for Go <1.16 compatibility)
type fileInfoWrapper struct {
	name  string
	isDir bool
}

func (f *fileInfoWrapper) Name() string {
	return f.name
}

func (f *fileInfoWrapper) Size() int64 {
	return 0
}

func (f *fileInfoWrapper) Mode() os.FileMode {
	return 0
}

func (f *fileInfoWrapper) ModTime() time.Time {
	return time.Time{}
}

func (f *fileInfoWrapper) IsDir() bool {
	return f.isDir
}

func (f *fileInfoWrapper) Sys() interface{} {
	return nil
}

// monitorBashHistory tracks changes in bash history files
func (p *Program) monitorBashHistory(agentID string) {
	cache := newEventCache()

	// Track the last read position for each history file
	lastReadPositions := make(map[string]int64)

	homeDir := "/home"
	
	for {
		// Get a fresh list of directories each time
		dirs, err := ioutil.ReadDir(homeDir)
		if err != nil {
			fmt.Printf("Failed to read home directories: %v\n", err)
			time.Sleep(30 * time.Second)
			continue
		}

		// Add root's home directory
		dirs = append(dirs, &fileInfoWrapper{name: "root", isDir: true})

		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}

			username := dir.Name()
			
			// Skip system users
			if strings.HasPrefix(username, "systemd-") || 
			   username == "nobody" || 
			   strings.HasPrefix(username, "_") {
				continue
			}
			
			var historyPath string

			if username == "root" {
				historyPath = "/root/.bash_history"
			} else {
				historyPath = filepath.Join(homeDir, username, ".bash_history")
			}

			// Check if the file exists and is readable
			fileInfo, err := os.Stat(historyPath)
			if err != nil {
				// File doesn't exist or can't be accessed
				continue
			}

			// Get the current file size
			currentSize := fileInfo.Size()

			// Get the last read position for this file
			lastPos, exists := lastReadPositions[historyPath]

			// If this is the first time we're reading this file,
			// or if the file hasn't changed, skip it
			if !exists {
				// First time seeing this file - just record its current size
				lastReadPositions[historyPath] = currentSize
				fmt.Printf("Tracking bash history for user %s (size: %d)\n", username, currentSize)
				continue
			}

			// If the file size hasn't increased, skip it
			if currentSize <= lastPos {
				continue
			}

			fmt.Printf("New bash history entries detected for user %s\n", username)

			// Open the file and seek to the last read position
			file, err := os.Open(historyPath)
			if err != nil {
				fmt.Printf("Failed to open bash history for user %s: %v\n", username, err)
				continue
			}

			// Seek to where we left off last time
			file.Seek(lastPos, 0)

			// Read only the new content
			newContent := make([]byte, currentSize-lastPos)
			_, err = file.Read(newContent)
			file.Close()

			if err != nil && err != io.EOF {
				fmt.Printf("Error reading bash history for user %s: %v\n", username, err)
				continue
			}

			// Update the last read position
			lastReadPositions[historyPath] = currentSize

			// Process each new line
			lines := strings.Split(string(newContent), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Create a unique key for this command
				cmdKey := fmt.Sprintf("%s_%s", username, line)
				if cache.shouldLog(cmdKey) {
					fmt.Printf("New bash history command for user %s: %s\n", username, line)
					
					activity := &TerminalActivity{
						Command:   line,
						User:      username,
						Timestamp: time.Now(),
					}

					log := Log{
						AgentID:        agentID,
						Timestamp:      time.Now().Format(time.RFC3339),
						Event:          "bash_history_command",
						Details:        fmt.Sprintf("New command executed: %s (user: %s)", line, username),
						Severity:       "low",
						AdditionalData: activity,
					}
					sendLog(log)
				}
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// Keeping these functions for compatibility with original code
func processcommand(line string) *TerminalActivity {
	// This is now a legacy function, but keeping for compatibility
	fmt.Printf("Legacy processcommand called with: %s\n", line)
	return nil
}

func setupAuditRules() {
	// This is now a legacy function, but keeping for compatibility
	fmt.Println("Legacy setupAuditRules called, but not doing anything")
}

// detectTerminal starts monitoring terminal commands
func (p *Program) detecterminal(agentID string) {
	// Start the process monitoring (primary method)
	go p.monitorProcessEvents(agentID)
	fmt.Println("Process monitoring started")

	// Also monitor bash history files for changes (secondary method)
	go p.monitorBashHistory(agentID)
	fmt.Println("Bash history monitoring started")
}

func sendLog(log Log) {
	logData, err := json.Marshal(log)
	if err != nil {
		fmt.Printf("Failed to serialize log: %v\n", err)
		return
	}

	resp, err := http.Post(backendURL, "application/json", bytes.NewBuffer(logData))
	if err != nil {
		fmt.Printf("Failed to send log: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if len(body) > 0 {
		fmt.Printf("Server response: %s\n", string(body))
	}
}


//Helper function
func sendOutput(out Output) {
	outData, err := json.Marshal(out)

	if err != nil {
		fmt.Printf("error_occured")
	}

	resp, err := http.Post(OutputURL, "application/json", bytes.NewBuffer(outData))

	if err != nil {
		fmt.Printf("failed to send output")

		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if len(body) > 0 {
		fmt.Printf("Server response: %s\n", string(body))
	}
}

// Function to send a single file to the server
func sendFileToServer(agentID, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add agent ID field
	err = writer.WriteField("agent_id", agentID)
	if err != nil {
		return fmt.Errorf("failed to write agent ID field: %v", err)
	}

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	// Copy file content to form
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	// Close writer
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "http://localhost:8080/upload", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "EDRAgent",
		DisplayName: "EDR Monitoring Agent",
		Description: "Comprehensive system security monitoring agent",
	}

	prg := &Program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Printf("Failed to initialize service: %v\n", err)
		return
	}

	mode := flag.String("mode", "service", "Mode to run the agent: service or normal")
	flag.Parse()

	if *mode == "normal" {
		prg.run()
		return
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err = s.Install()
			if err != nil {
				fmt.Printf("Failed to install service: %v\n", err)
			}
			return
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				fmt.Printf("Failed to uninstall service: %v\n", err)
			}
			return
		case "start":
			err = s.Start()

			if err != nil {
				fmt.Printf("Failed to start service: %v\n", err)
			}
			return
		case "stop":
			err = s.Stop()
			if err != nil {
				fmt.Printf("Failed to stop service: %v\n", err)
			}
			return
		}
	}

	err = s.Run()
	if err != nil {
		fmt.Printf("Service failed: %v\n", err)
	}
}
