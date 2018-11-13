// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//

package main

import (
    "log"
    "fmt"
    "time"
    "github.com/sparrc/go-ping"
)

// Ping host --------------------------------------------------------------------------
func PingProbe(config Config, result chan string, verbose bool) {
    // Get icmp from config
    for _,g := range config.Groups {
        for _, h := range g.Hosts {
    	    if h.Probe == "icmp" {
		go ExecutePinger(g.Name,h,config.Probes["icmp"],result,verbose)
	    }
	}
    }
}

// Create Pinger instance
func ExecutePinger(Group string, Host host, Probe probe, result chan string, verbose bool) {
    if verbose {
	log.Printf("ICMP: Start ping routine for group %s, host %s, retries %d",Group,Host.Fqdn,Probe.Retries)
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
    	    log.Printf("ICMP: Ping Host:%s Addr:%s PacketsSent:%d PacketsRecv:%d PacketLoss:%f",Host.Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	}
	// Print result to channel
	result <- fmt.Sprintf("%d %s %s %s %s %d %d %f %f %f %f",end,Group, Host.Probe, Host.Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss, stats.MinRtt.Seconds()*1000, stats.MaxRtt.Seconds()*1000, stats.AvgRtt.Seconds()*1000)
    
	// Check if execution time is smaller then step time	
	sleeptime := int64(Probe.Step)-exectime
	if sleeptime < 0 {
    	    log.Printf("ICMP: Step time %d is too small. Increase it",Probe.Step)
    	    sleeptime = 1
	}

	if verbose {
    	    log.Printf("ICMP: Ping host %s sleep %d seconds", Host.Fqdn,sleeptime)
	}
	
	time.Sleep(time.Duration(sleeptime) * time.Second)

    }
}

