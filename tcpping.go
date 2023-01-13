//
// Written by Naufan Rusyda Faikar <naufan.faikar@myrepublic.net.id>
//

package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Tcpping host --------------------------------------------------------------------------
func TcpPingProbe(config Config, result chan string, verbose bool) {
    var (
        records     int
        Ip          string
        PacketsSent int
        PacketsRecv int
        PacketsLoss float64
        MinRtt      float64
        MaxRtt      float64
        AvgRtt      float64
        ArrRtt      []string
    )

    // ip to fqdn and group mapper
    ip_fqdn := make(map[string]string)
    ip_group := make(map[string]string)
    // Tcpping args
    tcpping_args := config.Probes["tcpping"].Args

    // Get tcpping from config
    for _, g := range config.Groups {
        for _, h := range g.Hosts {
            if h.Probe == "tcpping" {
                ip_fqdn[h.Ip] = h.Fqdn
                ip_group[h.Ip] = g.Name
                tcpping_args = append(tcpping_args, h.Ip)
                records++
            }
        }
    }

    if verbose {
        log.Printf("TCPPING: Found %d tcpping records", records)
    }

    if records > 0 {
        if verbose {
            log.Printf("TCPPING: Running command:%s %v", config.Probes["tcpping"].Cmd, tcpping_args)
        }

        for {
            // Get start time
            start := time.Now().Unix()
            // Run Tcpping
            cmd := exec.Command(config.Probes["tcpping"].Cmd, tcpping_args...)
            stdout, err := cmd.StdoutPipe()
            if err != nil {
                log.Print(err)
            }
            if err := cmd.Start(); err != nil {
                log.Print(err)
            }

            // Get results
            buff := bufio.NewScanner(stdout)
            for buff.Scan() {
                text := buff.Text()
                if verbose {
                    log.Printf("TCPPING: %s\n", text)
                }

                outparsed := strings.Split(text, " : ")
                ArrRtt = strings.Split(outparsed[1], " ")
                Ip = outparsed[0]

                PacketsSent = len(ArrRtt)
                PacketsLoss = float64(strings.Count(outparsed[1], "-")) / float64(PacketsSent) * 100

                MinRtt, _ = strconv.ParseFloat(ArrRtt[0], 64)
                MaxRtt, _ = strconv.ParseFloat(ArrRtt[0], 64)

                var sum float64
                var count int

                for _, Rtt := range ArrRtt {
                    if Rtt == "-" {
                        continue
                    }
                    Rtt, _ := strconv.ParseFloat(Rtt, 64)
                    if Rtt < MinRtt {
                        MinRtt = Rtt
                    }
                    if Rtt > MaxRtt {
                        MaxRtt = Rtt
                    }
                    sum += Rtt
                    count += 1
                }

                AvgRtt = sum / float64(count)

                // Print result to channel
                result <- fmt.Sprintf("%d %s %s %s %s %d %d %f %f %f %f", time.Now().Unix(), ip_group[Ip], "tcpping", ip_fqdn[Ip], Ip, PacketsSent, PacketsRecv, PacketsLoss, MinRtt, MaxRtt, AvgRtt)
            }

            // Stop process
            cmd.Wait()

            // Eval execute time
            end := time.Now().Unix()
            exectime := end - start

            // Check if execution time is smaller then step time
            sleeptime := int64(config.Probes["tcpping"].Step) - exectime

            if sleeptime < 0 {
                log.Printf("TCPPING: Step time %d is too small. Increase it", config.Probes["tcpping"].Step)
                sleeptime = 1
            }

            if verbose {
                log.Printf("TCPPING: Sleep %d seconds", sleeptime)
            }

            time.Sleep(time.Duration(sleeptime) * time.Second)
        }
    }
}
