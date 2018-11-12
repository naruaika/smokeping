// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//
package main

// Import section ----------------------------------------------------------------------------------------

import (  
    "github.com/BurntSushi/toml"
    "fmt"
    "flag"
    "errors"
    "log"
    "os/signal"
    "os"
    "syscall"
//    "bufio"
//    "os/exec"
//    "strings"
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
    Host string
    Db string
    User string
    Pass string
    Step int
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
    Step int
    Cmd string
    Args []string
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
//-------------------------------------------------------------------------------------------------------
// Signal handler
func SignalHandler(verbose bool) {
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    if verbose {
        log.Print("Got exiting signal")
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
	log.Print(err)
	return
    }

// Create channels
    probe_output := make(chan string, 1024)

// Run database outputs
   switch config.Global_tags.Output {
    case "influx":
	go OutputInfluxDb(config.Global_tags.Project, config.Agent.Hostname, config.Database["influx"], probe_output, verbose)
    }

// Run probes
    for _,g := range config.Groups {
        for _, h := range g.Hosts {
	    probe , err:= GetProbe(config, h.Probe)
	    if err != nil {
		log.Printf("%s for host %s",err.Error(),h.Fqdn)
		continue
	    }
	    switch h.Probe {
		case "icmp":
	    	    go PingProbe(g.Name, h, probe, probe_output, verbose)
		case "fping":
		    go FPingProbe(g.Name, h, probe, probe_output, verbose)
	    }
	}
    }
    
//  Check system signals    
    SignalHandler(verbose)

// End
}
