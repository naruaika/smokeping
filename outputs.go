// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//

package main

import (
    "github.com/influxdata/influxdb/client/v2"
    "log"
    "fmt"
    "time"
)

//-------------------------------------------------------------------------------------------------------
// Get result from channel
func OutputInfluxDb(project string, agent_hostname string, scheme string, username string, password string, db string, measurement string, step int, result chan string, verbose bool) {

    var (
	unixtime int64
	group string
	fqdn string
	ip string
	sent int
	recv int
	loss float32
	min float32
	max float32
	avg float32
	batchpoint client.BatchPoints
	addpoints int
    )

    for {
	select {
	    case stats:= <- result:

		fmt.Sscanf(stats,"%d %s %s %s %d %d %f %f %f %f",&unixtime,&group,&fqdn,&ip,&sent,&recv,&loss,&min,&max,&avg)
	        // Create fields and tags
		tags:= map[string]string{
        	    "fqdn": fqdn,
		    "group": group,
		    "agent": agent_hostname,
		    "project": project,
		}
		fields := map[string]interface{}{
                    "loss": loss,
	            "min": min,
    	    	    "avg": avg,
    	    	    "max": max,
    		}

		// Create point
		pt,err := client.NewPoint(
    		    measurement,
    		    tags,
		    fields,
		    time.Unix(unixtime,0),
		)
    
		if err != nil {
		    log.Print(err)
		    break;
		}

		// Add point to batchpoint
		batchpoint.AddPoint(pt)
		addpoints ++
		if verbose {
			log.Printf("InfluxDb: %v",pt)
		}

	    default:
		
		// if we have addpoints
		if addpoints>0 {
		    if verbose {
			log.Printf("InfluxDb: Writing %d points",addpoints)
		    }
		
		    // Open InfluxDb connection		
		    infdbcon, err := client.NewHTTPClient(client.HTTPConfig{
    			Addr:     scheme,
    			Username: username,
    			Password: password,
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
		    Database: db,
		    Precision: "",
		})
		
		addpoints = 0

		if verbose {
		    log.Printf("InfluxDb: Sleep %d seconds", step)
		}

	        time.Sleep(time.Duration(step)*time.Second)    
	}
    }
}
//-------------------------------------------------------------------------------------------------------
