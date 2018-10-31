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
//    "time"
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
func PingHost(host string, retries int, packetsize int) {
    pinger, err := ping.NewPinger(host)
    if err != nil {
        log.Print(err)
	return
    }
    pinger.SetPrivileged(true)
    pinger.Count = retries
    pinger.Run() // blocks until finished
    stats := pinger.Statistics()
    fmt.Printf("%v",stats)
}

//-------------------------------------------------------------------------------------------------------

func main() { 

    var config Config

    configfile, _ := GetCommandLineArgs()
 
    if _, err := toml.DecodeFile(configfile, &config); err != nil {
	fmt.Println(err)
	return
    }

    for _,g := range config.Groups {
        for _, h := range g.Hosts {
	    probe , err:= GetProbe(config, h.Probe)
	    if err != nil {
		log.Printf("%s for host %s",err.Error(),h.Fqdn)
		continue
	    }
	    log.Printf("Probing host %s(%s) with probe %s\n", h.Fqdn, h.Ip, h.Probe)
	    
		if (h.Probe == "icmp") {
		    PingHost(h.Ip, probe.Retries, probe.Packet)
		}
	}
    }
}
