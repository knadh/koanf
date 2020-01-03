package main

import (
	"fmt"
	"log"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/s3"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	// Load JSON config from s3.
	if err := k.Load(s3.Provider(s3.Config{
		AccessKey: os.Getenv("AWS_S3_ACCESS_KEY"),
		SecretKey: os.Getenv("AWS_S3_SECRET_KEY"),
		Region:    os.Getenv("AWS_S3_REGION"),
		Bucket:    os.Getenv("AWS_S3_BUCKET"),
		ObjectKey: "mock/mock.json",
	}), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	fmt.Println("parent's name is = ", k.String("parent1.name"))
	fmt.Println("parent's ID is = ", k.Int("parent1.id"))
}
