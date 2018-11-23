// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//

package main

import (
    "log"
    "fmt"
    "time"
    "os/exec"
    "bufio"
    "strings"
)

// Fping host --------------------------------------------------------------------------
func FPingProbe(config Config, result chan string, verbose bool) {
    var (
	records int
        Ip string
        PacketsSent int
        PacketsRecv int
        PacketsLoss float64
        MinRtt float64
        MaxRtt float64
        AvgRtt float64
    )

    // ip to fqdn and group mapper
    ip_fqdn := make(map[string]string)
    ip_group := make(map[string]string)
    // Fping args
    fping_args := config.Probes["fping"].Args

    // Get fping from config
    for _,g := range config.Groups {
        for _, h := range g.Hosts {
    	    if h.Probe == "fping" {
		ip_fqdn[h.Ip] = h.Fqdn
		ip_group[h.Ip] = g.Name
		fping_args = append(fping_args, h.Ip)	    
		records ++
	    }
	}
    }
    if verbose {
	log.Printf("FPING: Found %d fping records",records)
    }
    if records>0 {
        if verbose {
        	log.Printf("FPING: Running command:%s %v", config.Probes["fping"].Cmd,fping_args)
	}
    	for {
    	    // Get start time
	    start := time.Now().Unix()
	    // Run Fping
	    cmd := exec.Command(config.Probes["fping"].Cmd, fping_args...)
	    stderr, err := cmd.StderrPipe()
	    if err != nil {
    		log.Print(err)
	    }
	    if err := cmd.Start(); err != nil {
    		log.Print(err)
	    }
	    
    	    // Get results
	    buff := bufio.NewScanner(stderr)
	    for buff.Scan() {
		text := buff.Text()
		if verbose {
		    log.Printf("FPING: %s\n",text)
		}
		// Regexp	    
		r := strings.NewReplacer(
		    " : xmt/rcv/%loss = ", " ",
        	    ", min/avg/max = ", " ",
		    "%","",
		    "/"," ",
		    "\n","",
		)
		outparsed := r.Replace(text)
		fmt.Sscanf(outparsed,"%s %d %d %f %f %f %f",&Ip,&PacketsSent,&PacketsRecv,&PacketsLoss,&MinRtt,&AvgRtt,&MaxRtt)	
    		// Print result to channel
		result <- fmt.Sprintf("%d %s %s %s %s %d %d %f %f %f %f",time.Now().Unix(), ip_group[Ip],"fping", ip_fqdn[Ip], Ip, PacketsSent, PacketsRecv, PacketsLoss, MinRtt, MaxRtt, AvgRtt)
	    }
	    // Stop process
	    cmd.Wait()

	    // Eval execute time
	    end := time.Now().Unix()
	    exectime := end - start

	    // Check if execution time is smaller then step time	
    	    sleeptime := int64(config.Probes["fping"].Step)-exectime

    	    if sleeptime < 0 {
		log.Printf("FPING: Step time %d is too small. Increase it", config.Probes["fping"].Step)
		sleeptime = 1
	    }

    	    if verbose {
    		log.Printf("FPING: Sleep %d seconds", sleeptime)
	    }

    	    time.Sleep(time.Duration(sleeptime) * time.Second)

	}
    
    }
}    	
