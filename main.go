package main

import (
	"kas/component-center/cc-etcd/lib"
	"log"
	"time"
)

func main() {
	cli, err := lib.NewEtcdHelper([]string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"})
	if err != nil {
		log.Fatal(err)
	}
	err = cli.Register("bc-good", "127.0.0.1", 8060, 3)
	if err != nil {
		log.Fatal(err)
	}

	<-time.After(20 * time.Second)
	cli.UnRegister()
}
