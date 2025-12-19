cat > main.go << 'EOF'
package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
å)

func main() {
	fails := 0
	for {
		time.Sleep(1000 * time.Millisecond)
		
		r, e := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if e != nil {
			fails++
			if fails > 2 {
				fmt.Println("Unable to fetch server statistic.")
			}
			continue
		}
		
		if r.StatusCode != 200 {
			r.Body.Close()
			fails++
			if fails > 2 {
				fmt.Println("Unable to fetch server statistic.")
			}
			continue
		}
		
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		
		fails = 0
		proc(string(b))
	}
}

func proc(d string) {
	v := strings.Split(strings.TrimSpace(d), ",")
	if len(v) != 6 {
		return
	}
	
	// 1
	if f, e := strconv.ParseFloat(v[0], 64); e == nil && f > 30 {
		fmt.Printf("Load Average is too high: %.2f\n", f)
	}
	
	// 2
	tm, _ := strconv.ParseUint(v[1], 10, 64)
	um, _ := strconv.ParseUint(v[2], 10, 64)
	if tm > 0 {
		mp := float64(um) * 100 / float64(tm)
		if mp > 80 {
			fmt.Printf("Memory usage too high: %.2f%%\n", mp)
		}
	}
	
	// 3
	td, _ := strconv.ParseUint(v[3], 10, 64)
	ud, _ := strconv.ParseUint(v[4], 10, 64)
	if td > 0 {
		dp := float64(ud) * 100 / float64(td)
		if dp > 90 {
			mb := float64(td-ud) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.2f Mb left\n", mb)
		}
	}
	
	// 4
	tb, _ := strconv.ParseUint(v[5], 10, 64)
	if tb > 0 {
		np := float64(um) * 100 / float64(tb)
		if np > 90 {
			mbits := float64(tb-um) * 8 / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", mbits)
		}
	}
}
EOF