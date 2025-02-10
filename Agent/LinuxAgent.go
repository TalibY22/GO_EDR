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
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
    "log"
    "strconv"
    "bufio"
    "os/user"
    "io"
    
    

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

    AgentID string `json:"agent_id"`
    Given_command string `json:"given_command"`
    Output string `json:"output"`
}

// NetworkConnection represents a detected network connection
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


//Terminal COmmands 
type TerminalActivity struct {
    Command     string
    Args        []string
    ProcessID   int
    User        string
    Timestamp   time.Time
    IsSudo      bool
    RealCommand string
    RealUser    string    // Original user when using sudo
}



type Command struct{
    Command string  `json:"command"`
    Arguments string `json:"arguments"`
}


type Response struct  {
    Data [] Command `json:"data"`
}



// Program represents the service
type Program struct {
    exit chan struct{}
}

const (
    backendURL = "http://localhost:8080/logs"
    OutputURL = "http://localhost:8080/output"
	min = 300 *time.Second
)


type eventCache struct {
    lastEvents map[string]time.Time
}


//Caching to prevent dupplicate logsv
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







//PROGRAM ENTRY POINT
func (p *Program) run() {
    agentID := generateAgentID()

    // Start all monitoring functions
    if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
        go monitorFiles(agentID, getDownloadsPath())
		fmt.Printf(getDownloadsPath())
        go monitorFiles(agentID, getDesktopPath())
     //  go p.monitorProcesses(agentID)
        //go p.monitorNetworkConnections(agentID)
        go p.monitorDNS(agentID)
        
        if runtime.GOOS == "windows" {
            //go p.monitorRegistry(agentID)
            go p.monitorPowerShell(agentID)
        }

        go p.Getcommand(agentID)
        
        go p.detecterminal(agentID)
        go p.monitorUSBDevices(agentID)
        go p.detectSuspiciousProcesses(agentID)
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
                        "pid":         proc.Pid,
                        "command":     cmdline,
                        "user":        username,
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

// PowerShell Monitoring Functions
func (p *Program) monitorPowerShell(agentID string) {
    if runtime.GOOS != "windows" {
        return
    }

    setupScript := `
    $LogPath = "$env:USERPROFILE\AppData\Local\PowerShellLogs"
    if (-not (Test-Path $LogPath)) {
        New-Item -ItemType Directory -Path $LogPath
    }
    Start-Transcript -Path "$LogPath\powershell_transcript.txt" -Append
    Enable-PSScriptBlockLogging
    `

    cmd := exec.Command("powershell", "-Command", setupScript)
    cmd.Run()

    suspiciousPatterns := []*regexp.Regexp{
        regexp.MustCompile(`(?i)invoke-expression`),
        regexp.MustCompile(`(?i)iex`),
        regexp.MustCompile(`(?i)downloadstring`),
        regexp.MustCompile(`(?i)bypass`),
        regexp.MustCompile(`(?i)encodedcommand`),
    }

    for {
        processes, _ := process.Processes()
        for _, proc := range processes {
            name, err := proc.Name()
            if err != nil {
                continue
            }

            if strings.ToLower(name) == "powershell.exe" {
                cmdline, _ := proc.Cmdline()
                username, _ := proc.Username()

                for _, pattern := range suspiciousPatterns {
                    if pattern.MatchString(cmdline) {
                        command := PowerShellCommand{
                            CommandLine: cmdline,
                            User:       username,
                            Timestamp:  time.Now(),
                            ProcessID:  proc.Pid,
                        }

                        log := Log{
                            AgentID:        agentID,
                            Timestamp:      time.Now().Format(time.RFC3339),
                            Event:          "suspicious_powershell",
                            Details:        fmt.Sprintf("Suspicious PowerShell command from user %s", username),
                            Severity:       "high",
                            AdditionalData: command,
                        }
                        sendLog(log)
                    }
                }
            }
        }
        time.Sleep(5 * time.Second)
    }
}

// Registry Monitoring Functions (Windows Only)


// Helper Functions
func generateAgentID() string {
    hostname, err := os.Hostname()
    if err != nil {
        return "unknown-agent"
    }
    return fmt.Sprintf("agent-%s", hostname)
}

func getDownloadsPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return ""
    }
    return filepath.Join(home, "Downloads")
}

func getDesktopPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return ""
    }
    return filepath.Join(home, "Desktop")
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


//Detect suspicious processes   
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



//Monitor is any usb devices are plugged in
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



//NOtify if a user logs in
func (p *Program) monitorUserBehavior(agentID string) {
    lastLoginCheck := time.Now()
    
   // if runtime.GOOS == "windows" {
     //   go windowslogin(agentID,lastLoginCheck)
   // }



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
    ticker := time.NewTicker(4 * time.Minute)
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
            var command Command
            if err := json.Unmarshal(body, &response); err != nil {
                log.Printf("Failed to unmarshal command: %v", err)
                continue
            }

            if len(response.Data) == 0 {
                continue // No commands to execute
            }

            command = response.Data[0]
            if command.Command == "" {
                log.Printf("Empty command received")
                continue
            }

            // Execute the command
            cmd := exec.Command(command.Command)
            output, err := cmd.CombinedOutput()
            if err != nil {
                log.Printf("Failed to execute command: %v", err)
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


// Check what is gonna be typed on the commandline
//audit rules to allow for checking of sudo commands 

func setupAuditRules() {
	rules := []string{
		"-a always,exit -F arch=b64 -S execve -k command_edr",                      // Monitor execve system calls
		"-w /usr/bin/sudo -p x -k command_edr", // Monitor sudo usage
	}

	for _, rule := range rules {
		cmd := exec.Command("auditctl", "-a", rule)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to set audit rule '%s': %v\n", rule, err)
		}
	}
}



func removeEmpty(s []string) []string {
    var r []string
    for _, str := range s {
        if str != "" {
            r = append(r, str)
        }
    }
    return r
}


//Make the terminal command nicer 
func enrichCommandInfo(activity *TerminalActivity) {

    cmdline,err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline",activity.ProcessID))

    if err != nil {
        fmt.Printf("Failed to read command line for PID %d: %v\n", activity.ProcessID, err)
        return
    }


    //Split the command line into parts
   parts := strings.Split(string(cmdline), "\x00")

   parts = removeEmpty(parts)


   if len(parts) == 0{
    return
   }

   if parts[0] == "sudo" {
    activity.IsSudo = true
    activity.RealUser = activity.User
    activity.User = "root"
    activity.RealCommand = parts[1]

    activity.Args = parts[2:]

   
    
    
    } else {
        activity.Command = parts[0]
        print(activity.Command)
        activity.Args = parts[1:]
    }


    environ,_ := os.ReadFile(fmt.Sprintf("/proc/%d/environ",activity.ProcessID))

    env := strings.Split(string(environ), "\x00")

    for _,e := range env {
        if strings.HasPrefix(e,"SUDO_USER=") {
            activity.RealUser = strings.TrimPrefix(e,"SUDO_USER=")
            break
        }
    }
}





//Process the command executed 
func processcommand(line string) *TerminalActivity  {


    if !strings.Contains(line, "SYSCALL") && !strings.Contains(line,"PATH") {

        return nil 
    }



    activity := &TerminalActivity{

        Timestamp: time.Now(),
    }


    fields := strings.Fields(line)

    for  _, field := range fields {


        switch {

            case strings.HasPrefix(field,"exe="):
                
                activity.Command = strings.Trim(field[4:], "\"")
            case strings.HasPrefix(field,"pid="):
                activity.ProcessID,_ = strconv.Atoi(field[4:])

            case strings.HasPrefix(field,"uid="):
                //Get username from uid 
                uid := field[4:]
                if _, err := strconv.Atoi(uid); err != nil {
                fmt.Printf("Invalid UID: %s, not a numeric value\n", uid)
                activity.User = uid
                } else {
                   u, err := user.LookupId(uid)
                   if err != nil {
                    fmt.Printf("Failed to get username for uid %s: %v\n", uid, err)
                   activity.User = uid
                      } else {
                         activity.User = u.Username
    }
}

                


        }


    }

    if activity.ProcessID > 0 {
        enrichCommandInfo(activity)
    }

    return activity
}






func (p *Program) detecterminal (agentID string) {


    setupAuditRules()

    cmd := exec.Command("ausearch", "-k", "command_edr", "--start", "recent")




    stdout,err := cmd.StdoutPipe()

    if err != nil {

        fmt.Printf("erro starting monitoing of the terminal")
        return
    }

    if err := cmd.Start(); err != nil {

        fmt.Printf("issue running the command")
        return
    }

    scanner := bufio.NewScanner(stdout)



    for scanner.Scan() {
        if activity := processcommand(scanner.Text()); activity != nil {
            log := Log{
                AgentID:   agentID,
                Timestamp: time.Now().Format(time.RFC3339),
                Event:     "terminal_command",
                Details:   fmt.Sprintf("Terminal command executed: %s", activity.Command),
                Severity:  "idk",
                AdditionalData: activity,
            }
            sendLog(log)
        }
    }

} 








//Send data to api
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



func sendOutput(out Output) {
    outData,err := json.Marshal(out)

    if err != nil {
        fmt.Printf("error_occured")
    }

    resp,err := http.Post(OutputURL, "application/json",bytes.NewBuffer(outData))

    if err != nil {
        fmt.Printf("failed to send output")

        return 
    }

    defer resp.Body.Close()

    body,_ := ioutil.ReadAll(resp.Body)

    if len(body) > 0 {
        fmt.Printf("Server response: %s\n",string(body))
    }
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



