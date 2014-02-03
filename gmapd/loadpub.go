package main

import (
	"fmt"
	. "github.com/peak6/logger"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func dumpStats() {
	i1, t1 := getCPUSample()
	time.Sleep(1 * time.Second)
	i2, t2 := getCPUSample()

	Linfo.Println("Sample:", i2-i1, t2-t1)

}
func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}
