package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/knadh/koanf/providers/k8smount"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

func main() {
	p := k8smount.Provider("mock/mount", ".", k8smount.Opt{
		TransformFunc: func(k, v string) (string, any) {
			return strings.ToLower(strings.ReplaceAll(k, "_", ".")), strings.TrimSpace(v)
		},
	})

	if err := k.Load(p, nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("database's host is = ", k.String("database.host"))
	fmt.Println("database's port is = ", k.Int("database.port"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := p.Watch(func(_ any, err error) {
		if err != nil {
			log.Printf("watch error: %v", err)
			return
		}

		log.Println("config changed. Reloading ...")
		k.Load(p, nil)
		k.Print()
	}); err != nil {
		log.Fatalf("error watching config: %v", err)
	}

	log.Println("waiting forever. Try making a change under mock/mount/ to live reload")
	<-ctx.Done()
}
