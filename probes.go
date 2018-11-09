// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//

package main

import (
    "log"
    "fmt"
    "time"
    "os/exec"
    "strings"
    "github.com/sparrc/go-ping"
)

// ping host --------------------------------------------------------------------------
func PingProbe(Group string, Host host, Probe probe, result chan string, verbose bool) {
    if verbose {
	log.Printf("Probe: Start ping routine for group %s, host %s, retries %d",Group,Host.Fqdn,Probe.Retries)
    }
    
    for {
	
	pinger, err := ping.NewPinger(Host.Ip)
        if err != nil {
	    log.Print(err)
	    return
	}

	pinger.SetPrivileged(true)
	pinger.Count = Probe.Retries
        start := time.Now().Unix()

    	pinger.Run() // blocks until finished
	end := time.Now().Unix()
	exectime := end - start

	stats := pinger.Statistics()
	pinger = nil
	if verbose {	
	    log.Printf("Probe: Ping Host:%s Addr:%s PacketsSent:%d PacketsRecv:%d PacketLoss:%f",Host.Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	}
	// Print result to channel
	result <- fmt.Sprintf("%d %s %s %s %s %d %d %f %f %f %f",end,Group, Host.Probe, Host.Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss, stats.MinRtt.Seconds()*1000, stats.MaxRtt.Seconds()*1000, stats.AvgRtt.Seconds()*1000)
	
	// Check if execution time is smaller then step time	
	sleeptime := int64(Probe.Step)-exectime
	if sleeptime < 0 {
	    log.Printf("Probe: Step time %d is too small. Increase it",Probe.Step)
	    sleeptime = 1
	}

	if verbose {
	    log.Printf("Probe: Ping host %s sleep %d seconds", Host.Fqdn,sleeptime)
	}
        time.Sleep(time.Duration(sleeptime) * time.Second)

    }
}

// Fping host --------------------------------------------------------------------------
func FPingProbe(Group string, Host host, Probe probe, result chan string, verbose bool) {
    var (
	Ip string
	PacketsSent int
	PacketsRecv int
	PacketsLoss float64
	MinRtt float64
	MaxRtt float64
	AvgRtt float64
    )

    arg := append(Probe.Args, Host.Ip)
    if verbose {
	log.Printf("Probe: Fping group %s, host %s start command %s, args %v",Group,Host.Fqdn,Probe.Cmd,arg)
    }

    for {
	// Get start time
	start := time.Now().Unix()
	// Run Fping
	cmd := exec.Command(Probe.Cmd,arg...)
	out,_ := cmd.CombinedOutput()
	if verbose {	
	    log.Printf("Probe: Fping group %s, host %s: cmd out %s",Group, Host.Fqdn, out)
	}
	// Get results
	r := strings.NewReplacer(
	    " : xmt/rcv/%loss = ", " ",
            ", min/avg/max = ", " ",
	    "%","",
	    "/"," ",
	    "\n","",
	)
	outstr := r.Replace(string(out[:]))
	fmt.Sscanf(outstr,"%s %d %d %f %f %f %f",&Ip,&PacketsSent,&PacketsRecv,&PacketsLoss,&MinRtt,&AvgRtt,&MaxRtt)	
	// Get end time
	end := time.Now().Unix()
	exectime := end - start
	// Print result to channel
	result <- fmt.Sprintf("%d %s %s %s %s %d %d %f %f %f %f",end, Group, Host.Probe, Host.Fqdn, Ip, PacketsSent, PacketsRecv, PacketsLoss, MinRtt, MaxRtt, AvgRtt)

	// Check if execution time is smaller then step time	
	sleeptime := int64(Probe.Step)-exectime
	if sleeptime < 0 {
	    log.Printf("Probe: Fping group %s, host %s step time %d is too small. Increase it",Group, Host.Fqdn, Probe.Step)
	    sleeptime = 1
	}

	if verbose {
	    log.Printf("Probe: Fping group %s, host %s sleep %d seconds", Group, Host.Fqdn, sleeptime)
	}
        time.Sleep(time.Duration(sleeptime) * time.Second)

    }
}
