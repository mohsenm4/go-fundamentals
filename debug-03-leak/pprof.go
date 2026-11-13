package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
	}()
	go func() {
		for {
			time.Sleep(3 * time.Second)
			log.Printf("goroutines=%d", runtime.NumGoroutine())
		}
	}()
}
