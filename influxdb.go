//
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

// -------------------------------------------------------------------------------------------------------
// Get result from channel
func OutputInfluxDb(project string, agent_hostname string, influxdb databaseConnector, result chan string, verbose bool) {

    var (
        unixtime    int64
        group       string
        measurement string
        fqdn        string
        ip          string
        sent        int
        recv        int
        loss        float64
        min         float64
        max         float64
        avg         float64
        batchpoint  client.BatchPoints
        addpoints   int
    )

    for {
        select {
            case stats := <-result:
                fmt.Sscanf(stats, "%d %s %s %s %s %d %d %f %f %f %f", &unixtime, &group, &measurement, &fqdn, &ip, &sent, &recv, &loss, &min, &max, &avg)

                if verbose {
                    log.Printf("INFLUXDB: Receive %s", stats)
                }

                // Create fields and tags
                tags := map[string]string{
                    "fqdn":    fqdn,
                    "group":   group,
                    "agent":   agent_hostname,
                    "project": project,
                }
                fields := map[string]interface{}{
                    "loss": loss,
                    "min":  min,
                    "avg":  avg,
                    "max":  max,
                }

                // Create point
                pt, err := client.NewPoint(
                    measurement,
                    tags,
                    fields,
                    time.Unix(unixtime, 0),
                )

                if err != nil {
                    log.Print(err)
                    break
                }

                // Add point to batchpoint
                batchpoint.AddPoint(pt)
                addpoints++
                if verbose {
                    log.Printf("INFLUXDB: [%s] Batchpoint %v", time.Unix(unixtime, 0).String(), pt)
                }

            default:
                // if we have addpoints
                if addpoints > 0 {
                    if verbose {
                        log.Printf("INFLUXDB: Writing %d points", addpoints)
                    }

                    // Open InfluxDb connection
                    infdbcon, err := client.NewHTTPClient(client.HTTPConfig{
                        Addr:     influxdb.Host,
                        Username: influxdb.User,
                        Password: influxdb.Pass,
                    })

                    if err != nil {
                        log.Print(err)
                    } else {
                        // Write batchpoint to influx database
                        if err := infdbcon.Write(batchpoint); err != nil {
                            log.Print(err)
                        }

                        // Close client resources
                        if err := infdbcon.Close(); err != nil {
                            log.Print(err)
                        }
                    }
                }

                // Generate new batchpoint structure
                batchpoint, _ = client.NewBatchPoints(client.BatchPointsConfig{
                    Database:  influxdb.Db,
                    Precision: "ms",
                })

                addpoints = 0

                if verbose {
                    log.Printf("INFLUXDB: Sleep %d seconds", influxdb.Step)
                }

                time.Sleep(time.Duration(influxdb.Step) * time.Second)
        }
    }
}

//-------------------------------------------------------------------------------------------------------
