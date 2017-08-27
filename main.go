package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

var targetRegexp = regexp.MustCompile(`gs:\/\/([^\/]+)/(.*)`)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		opt := parseOptions(r)
		ctx := r.Context()

		client, err := storage.NewClient(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fmt.Sprintf("%s.zip", opt.PathPrefix)))

		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		bucket := client.Bucket(opt.BucketName)
		it := bucket.Objects(ctx, &storage.Query{Prefix: opt.PathPrefix})
		for {
			obj, err := it.Next()
			if err != nil {
				if err == iterator.Done {
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			log.Printf("handling file %v", obj.Name)

			r, err := bucket.Object(obj.Name).NewReader(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			wr, err := zipWriter.Create(obj.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			if _, err := io.Copy(wr, r); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			log.Printf("done handling file: %v", obj.Name)
		}
	})

	fmt.Println("server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func parseOptions(r *http.Request) (opt Options) {
	if opt.Target = r.URL.Query().Get("uri"); opt.Target != "" {
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
