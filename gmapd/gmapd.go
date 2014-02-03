package main

import (
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/peak6/gmap"
	. "github.com/peak6/logger"
	"strings"
	"time"
)

var listen string
var join string
var myNode string

func init() {
	flag.StringVar(&listen, "l", ":0", "Listen port")
	flag.StringVar(&join, "j", "", "Join node")
	flag.StringVar(&myNode, "n", "", "Override node name")
}

func main() {
	flag.Parse()
	InitLogger()
	if myNode != "" {
		gmap.MyNode.Name = myNode
	}

	var nodesToJoin []string
	if join != "" {
		gmap.MyStore.PutStatic("/services/bar", "slave2")
		nodesToJoin = strings.Split(join, ",")
	} else {
		gmap.MyStore.PutStatic("/services/foo1", "master1")
		gmap.MyStore.PutStatic("/services/foo2", "master2")
	}
	go dumper()
	dumpStats()
	// Linfo.Println(gmap.MyStore.Spew())
	gmap.StartAutoDiscover()
	err := gmap.ListenAndJoin(listen, nodesToJoin)
	if err != nil {
		Lerr.Println("Failed to start:", err)
	}
}

func dumper() {
	c := time.Tick(5 * time.Second)
	for _ = range c {
		Linfo.Printf("DUMP:\n%s", spew.Sprint(gmap.MyStore))
		dumpStats()
	}
}
