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

// ping host --------------------------------------------------------------------------
func PingProbe(Group string, Fqdn string, Ip string , Retries int, Size int,Step int,result chan string, verbose bool) {
    if verbose {
	log.Printf("Probe: Start ping routine for group %s, host %s, retries %d, packet size %d",Group,Fqdn,Retries,Size)
    }
    
    for {
	
	pinger, err := ping.NewPinger(Ip)
        if err != nil {
	    log.Print(err)
	    return
	}

	pinger.SetPrivileged(true)
	pinger.Count = Retries
	pinger.Size = Size

        start := time.Now().Unix()
    	pinger.Run() // blocks until finished
	end := time.Now().Unix()
	exectime := end - start

	stats := pinger.Statistics()
	pinger = nil
	if verbose {	
	    log.Printf("Probe: Host:%s Addr:%s PacketsSent:%d PacketsRecv:%d PacketLoss:%f",Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	}
	// Print result to channel
	result <- fmt.Sprintf("%d %s %s %s %d %d %f %f %f %f",end,Group,Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss, stats.MinRtt.Seconds()*1000, stats.MaxRtt.Seconds()*1000, stats.AvgRtt.Seconds()*1000)
	
	// Check if execution time is smaller then step time	
	sleeptime := int64(Step)-exectime
	if sleeptime < 0 {
	    log.Printf("Probe: Step time %d is too small. Increase it",Step)
	    sleeptime = 1
	}

	if verbose {
	    log.Printf("Probe: Ping host %s sleep %d seconds", Fqdn,sleeptime)
	}
        time.Sleep(time.Duration(sleeptime) * time.Second)

    }
}
