//WIndowsssssssss







func windowslogin(agentid string,lastcheck time.Time) {


    evtlog,err := eventlog.Open("Security")

    if err != nil {
        fmt.Printf("eroor")
    }

    defer evtlog.Close()

    records ,err := evtlog.Read(eventlog.Backwards,0)

    if err != nil {
        fmt.Printf("issue readinf the event log ")
    }

    for _,record := range records {

    
        if strings.Contains(record.String(),"Event ID: 4624") {

  
             log := Log{
                AgentID:   agentid,
                Timestamp: time.Now().Format(time.RFC3339),
                Event:     "user_login",
                Details:   fmt.Sprintf("Windows login detected: %s", record.String()),
                //Severity:  "medium",
            }
            sendLog(log)
        }
    }



}



