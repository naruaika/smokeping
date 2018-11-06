// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//  
package main

// Import section ----------------------------------------------------------------------------------------

import (  
//    "github.com/influxdata/influxdb/client/v2"
    "github.com/BurntSushi/toml"
    "github.com/sparrc/go-ping"
    "fmt"
    "flag"
    "errors"
    "log"
//    "os"
//    "bufio"
//    "os/exec"
//    "strings"
    "time"
//    "strconv"
)
// Usage ------------------------------------------------------------------------------------------------
var usage = `
Usage:

    go-smokeping [-config configfile.toml] [-verbose]
`

// Config file structure --------------------------------------------------------------------------------
type Config struct {
    Global_tags globalTags
    Agent agentInfo
    Database map[string] databaseConnector
    Probes map[string] probe `toml:"probe"`
    Groups []group `toml:"group"`
}
 
type globalTags struct {
    Project string
    Output string
}

type agentInfo struct {
    Hostname string
}

type databaseConnector struct {
    Type string
    Host string
    Db string
    Measurement string
    User string
    Pass string
}
  
type host struct {
    Ip string
    Fqdn string
    Probe string
}

type group struct {
    Name string
    Hosts []host `toml:"host"`
}

type probe struct {
    Retries int
    Packet int
    Step int64
}


// Get command line parameters --------------------------------------------------------------------------
func GetCommandLineArgs() (string,bool) {
    var (
	configfile string
	verbose bool
    )
// Get command line args
    flag.StringVar(&configfile, "config","go-smokeping.toml","Config file locaion")
    flag.BoolVar(&verbose, "verbose",false,"true/false")
    flag.Usage = func() {
	fmt.Printf(usage)
    }
    flag.Parse()

    return configfile, verbose
}

// Return probe by name ---------------------------------------------------------------
func GetProbe(config Config, probe_name string) (probe, error) {
    var(
	 pr probe
	 ok bool
    )

    if pr, ok = config.Probes[probe_name]; ok {
	return pr, nil
    } else {
	return pr,errors.New("Probe "+probe_name+" not found")
    }
}

// ping host --------------------------------------------------------------------------
func PingHost(h host,p probe, result chan string, verbose bool) {
    for {
	pinger, err := ping.NewPinger(h.Ip)
	    if err != nil {
    		log.Print(err)
		return
	    }
	pinger.SetPrivileged(true)
	pinger.Count = p.Retries
	pinger.Size = p.Packet

        start := time.Now().Unix()
    	pinger.Run() // blocks until finished

	end := time.Now().Unix()
	exectime := end - start
	if verbose {
	    log.Printf("Host %s - start time %d, stop time:%d, cycle time:%d",h.Fqdn,start,end,exectime)
	}

	stats := pinger.Statistics()
	if verbose {	
	    log.Printf("Result: Host:%s Addr:%s PacketsSent:%d PacketsRecv:%d PacketLoss:%f",h.Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	}
	// Print result to channel
	result <- fmt.Sprintf("Unixtime:%d Host:%s Ip:%s PacketsSent:%d PacketsRecv:%d PacketLoss:%f MinRtt:%f ms MaxRtt:%f ms AvgRtt:%f ms",end, h.Fqdn, stats.Addr, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss, stats.MinRtt.Seconds()*1000, stats.MaxRtt.Seconds()*1000, stats.AvgRtt.Seconds()*1000)

	// Check if execution time is smaller then step time	
	sleeptime := p.Step-exectime
	if sleeptime < 0 {
	    log.Printf("Step time %d is too small. Increase it",p.Step)
	    sleeptime = 1
	}

	if verbose {
	    log.Printf("Sleep %d seconds", sleeptime)
	}
        time.Sleep(time.Duration(sleeptime) * time.Second)

    }
}

//-------------------------------------------------------------------------------------------------------
// Get result from channel
func GetResultFromChannel(result chan string, verbose bool) {
    for {
	stats := <- result
	fmt.Printf("Result: %v\n",stats)
	time.Sleep(10*time.Second)    
    }
}
//-------------------------------------------------------------------------------------------------------

func main() { 

    var (
      config Config
      verbose bool
    )

// Get command line arguments
    configfile, verbose := GetCommandLineArgs()
 
// Read config file
    if _, err := toml.DecodeFile(configfile, &config); err != nil {
	fmt.Println(err)
	return
    }

// Run pinger instances
    pinger_output := make(chan string, 1024)

    for _,g := range config.Groups {
        for _, h := range g.Hosts {
	    probe , err:= GetProbe(config, h.Probe)
	    if err != nil {
		log.Printf("%s for host %s",err.Error(),h.Fqdn)
		continue
	    }
	
	    log.Printf("Probing host %s(%s) with probe %s\n", h.Fqdn, h.Ip, h.Probe)
	    
	    if (h.Probe == "icmp") {
	        go PingHost(h, probe, pinger_output, verbose)
	    }
	}
    }
    
    GetResultFromChannel(pinger_output,verbose)
// End
}
