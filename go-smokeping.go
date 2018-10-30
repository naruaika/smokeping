// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>

package main

// Import section ----------------------------------------------------------------------------------------

import (  
//    "github.com/influxdata/influxdb/client/v2"
    "github.com/BurntSushi/toml"
    "fmt"
    "flag"
//    "log"
//    "os"
//    "bufio"
//    "os/exec"
//    "strings"
//    "time"
//    "strconv"
)

// Config file structure --------------------------------------------------------------------------------
type Config struct {
    Global_tags globalTags
    Agent agentInfo
    Database databaseConnector
    Probes []probe `toml:"probe"`
    Groups []group `toml:"group"`
}
 
type globalTags struct {
    Project string
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
}

type group struct {
    Name string
    Hosts []host `toml:"host"`
}

type probe struct {
    Name string
    Retries int
    Packet int
}


// Get command line parameters

func GetCommandLineArgs() (string,bool) {
    var (
	configfile string
	verbose bool
    )
// Get command line args

    flag.StringVar(&configfile, "config","go-smokeping.toml","Config file locaion")
    flag.BoolVar(&verbose, "verbose",false,"true/false")
    flag.Parse()
    return configfile, verbose
}
//-------------------------------------------------------------------------------------------------------

func main() { 

    var config Config

    configfile, _ := GetCommandLineArgs()
 
    if _, err := toml.DecodeFile(configfile, &config); err != nil {
	fmt.Println(err)
	return
    }

    fmt.Printf("Connector:%s %s\n", config.Database.Type ,config.Database.Host)
    for _,g := range config.Groups {
        fmt.Printf("Group: %v\n", g.Name)
        for _, h := range g.Hosts {
	    fmt.Printf("\tHost %s - %s\n", h.Fqdn, h.Ip)
	}
    }
}
