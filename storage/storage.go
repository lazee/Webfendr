package storage

import (
	"archive/zip"
	"cloud.google.com/go/storage"
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"webfendr/config"
)

// Create new Google Storage client
func createStorageClient(ctx context.Context) (*storage.Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func downloadAndUnpack(ctx context.Context, cfg *config.Config, object string, etag string, bucket *storage.BucketHandle) {
	path := cfg.SiteDir
	zipFile := filepath.Join(path, object)
	if !shouldDownload(zipFile, etag) {
		return
	}

	rc, err := bucket.Object(object).NewReader(ctx)
	if err != nil {
		log.Error("Got error")
		return
	}

	defer func(rc *storage.Reader) {
		err := rc.Close()
		if err != nil {
			log.Warning("Not able to close reader when downloading zip. Look into if this are causing any memory issues.")
		}
	}(rc)

	err, done := prepareDir(path, object, err)
	if done {
		log.Error("Something went wrong when preparing dir")
		return
	}

	f, err := os.Create(zipFile)
	if err != nil {
		log.Error("Could not create file " + f.Name())
		return
	}

	if _, err := io.Copy(f, rc); err != nil {
		log.Error("Could not write data to file " + f.Name())
		return
	}

	if err = f.Close(); err != nil {
		log.Error("Could not close file " + f.Name())
		return
	}
	unzipFile(zipFile, path, object)
	writeEtag(zipFile, etag)
}

func unzipFile(zipFile string, path string, object string) {
	dst := createObjectDirPath(path, object)
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Error(err)
		return
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		log.Debug("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			log.Error("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				log.Error("Could not create dir")
				return
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Error(err)
			return
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Error(err)
			return
		}

		fileInArchive, err := f.Open()
		if err != nil {
			log.Error(err)
			return
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			log.Error(err)
			return
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func shouldDownload(path string, etag string) bool {
	if _, err := os.Stat(path); err == nil {
		etagFile := path + ".etag"
		if _, err := os.Stat(etagFile); err == nil {
			content, err := os.ReadFile(etagFile)
			if err != nil {
				return true
			}
			if strings.TrimSpace(string(content)) == etag {
				return false
			}
		}
	}
	return true
}

func writeEtag(path string, etag string) {
	f, err := os.Create(path + ".etag")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(etag)
	if err2 != nil {
		log.Fatal(err2)
	}
}

func createObjectDirPath(path string, object string) string {
	fileName := filepath.Base(object)
	return filepath.Join(path, filepath.Dir(object), strings.TrimSuffix(fileName, filepath.Ext(fileName)))
}

func prepareDir(path string, object string, err error) (error, bool) {
	dir := createObjectDirPath(path, object)
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		err := os.RemoveAll(dir)
		if err != nil {
			return nil, true
		}
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, true
	}
	return err, false
}

func run(ctx context.Context, cfg *config.Config, bucket *storage.BucketHandle) error {
	it := bucket.Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if attrs.Size > 0 && strings.HasSuffix(attrs.Name, ".zip") {
			downloadAndUnpack(ctx, cfg, attrs.Name, attrs.Etag, bucket)
		}
	}

	return nil
}

func Syncer(ctx context.Context, cfg *config.Config) {
	log.Info("Starting Google Storage sync routine")

	cli, err := createStorageClient(ctx)
	if err != nil {
		log.Panicf("Could not create Google Client")
	}
	defer cli.Close()
	if cfg.GoogleStorageSync && cfg.GoogleStorageSyncBucket == "" {
		log.Panicf("Missing GOOGLE_STORAGE_BUCKET")
	}
	bucket := cli.Bucket(cfg.GoogleStorageSyncBucket)

	// Run initially at startup before ticking
	err = run(ctx, cfg, bucket)
	if err != nil {
		log.Error(err)
	}

	// Start ticking
	t := time.Tick(time.Duration(cfg.GoogleStorageSyncInterval) * time.Second)
	for {
		select {
		case <-t:
			err = run(ctx, cfg, bucket)
			if err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
