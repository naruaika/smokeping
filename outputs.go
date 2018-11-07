// 
// Written by Romanenko Denys <romanenkodenys@gmail.com>
//

package main

import (
//    "log"
    "fmt"
    "time"
)

//-------------------------------------------------------------------------------------------------------
// Get result from channel
func OutputInfluxDb(result chan string, verbose bool) {
    for {
	select {
	    case stats:= <- result:
		fmt.Printf("Result: %v\n",stats)
	    default:
	        time.Sleep(10*time.Second)    
	}
    }
}
//-------------------------------------------------------------------------------------------------------
