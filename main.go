package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

var targetRegexp = regexp.MustCompile(`gs:\/\/([^\/]+)/(.*)`)

func main() {
	opt := parseOptions()
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		exitWithError(err)
	}

	outputZipFile, err := os.Create("output.zip")
	if err != nil {
		exitWithError(err)
	}
	defer outputZipFile.Close()

	zipWriter := zip.NewWriter(outputZipFile)
	defer zipWriter.Close()

	bucket := client.Bucket(opt.BucketName)
	it := bucket.Objects(ctx, &storage.Query{Prefix: opt.PathPrefix})
	for {
		obj, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				return
			}
			exitWithError(err)
		}

		log.Printf("handling file %v", obj.Name)

		r, err := bucket.Object(obj.Name).NewReader(ctx)
		if err != nil {
			exitWithError(err)
		}

		w, err := zipWriter.Create(obj.Name)
		if err != nil {
			exitWithError(err)
		}

		if _, err := io.Copy(w, r); err != nil {
			exitWithError(err)
		}

		log.Printf("done handling file: %v", obj.Name)
	}

}

func parseOptions() (opt Options) {
	flag.Parse()
	if opt.Target = flag.Arg(0); opt.Target != "" {
		matches := targetRegexp.FindStringSubmatch(opt.Target)
		if len(matches) == 3 {
			opt.BucketName = matches[1]
			opt.PathPrefix = matches[2]
		}
	}
	return
}

type Options struct {
	Target     string
	BucketName string
	PathPrefix string
}

func exitWithError(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}
