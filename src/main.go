package main

import (
	"flag"
	"fmt"
	"github.com/qiniu/api.v6/auth/digest"
	"github.com/qiniu/api.v6/conf"
	"os"
	"qfetch"
	"runtime"
	"strings"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var job string
	var worker int
	var file string
	var bucket string
	var accessKey string
	var secretKey string
	var logFile string
	var checkExists bool

	//support qos, should be specified both
	var rsHost string
	var ioHost string

	flag.Usage = func() {
		fmt.Println(`Usage of qfetch:
  -ak="": qiniu access key
  -sk="": qiniu secret key
  -bucket="": qiniu bucket
  -job="": job name to record the progress
  -file="": resource list file to fetch
  -worker=0: max goroutine in a worker group
  -check-exists: check whether file exists in bucket
  -log="": fetch runtime log file
  -rs-host="": rs host to support specified qos system
  -io-host="": io host to support specified qos system version 2.1`)
	}

	flag.StringVar(&job, "job", "", "job name to record the progress")
	flag.IntVar(&worker, "worker", 1, "max goroutine in a worker group")
	flag.StringVar(&file, "file", "", "resource file to fetch")
	flag.StringVar(&bucket, "bucket", "", "qiniu bucket")
	flag.StringVar(&accessKey, "ak", "", "qiniu access key")
	flag.StringVar(&secretKey, "sk", "", "qiniu secret key")
	flag.StringVar(&logFile, "log", "", "fetch runtime log file")
	flag.BoolVar(&checkExists, "check-exists", false, "check whether file exists in bucket")
	flag.StringVar(&rsHost, "rs-host", "", "rs host to support specified qos system")
	flag.StringVar(&ioHost, "io-host", "", "io host to support specified qos system")

	flag.Parse()

	if accessKey == "" {
		fmt.Println("Error: accessKey is not set")
		return
	}

	if secretKey == "" {
		fmt.Println("Error: secretKey is not set")
		return
	}

	if bucket == "" {
		fmt.Println("Error: bucket is not set")
		return
	}

	if job == "" {
		fmt.Println("Error: job name is not set")
		return
	}

	if file == "" {
		fmt.Println("Error: resource file to fetch not set")
		return
	}
	_, ferr := os.Stat(file)
	if ferr != nil {
		fmt.Println(fmt.Sprintf("Error: file '%s' not exist", file))
		return
	}

	if worker <= 0 {
		fmt.Println("Error: worker count must larger than zero")
		return
	}

	if (rsHost != "" && ioHost == "") || (rsHost == "" && ioHost != "") {
		fmt.Println("Error: rs host and io host should be specified together")
		return
	}

	mac := digest.Mac{
		AccessKey: accessKey,
		SecretKey: []byte(secretKey),
	}

	if rsHost != "" && ioHost != "" {
		conf.IO_HOST = ioHost
		conf.RS_HOST = rsHost
	} else {
		//get bucket info
		bucketInfo, gErr := qfetch.GetBucketInfo(&mac, bucket)
		if gErr != nil {
			fmt.Println("Error: get bucket info error", gErr)
			return
		} else {
			switch bucketInfo.Region {
			case "z0":
				conf.RS_HOST = "http://rs.qbox.me"
				conf.IO_HOST = "http://iovip.qbox.me"
			case "z1":
				conf.RS_HOST = "http://rs-z1.qbox.me"
				conf.IO_HOST = "http://iovip-z1.qbox.me"
			case "z2":
				conf.RS_HOST = "http://rs-z2.qbox.me"
				conf.IO_HOST = "http://iovip-z2.qbox.me"
			case "na0":
				conf.RS_HOST = "http://rs-na0.qbox.me"
				conf.IO_HOST = "http://iovip-na0.qbox.me"
			case "as0":
				conf.RS_HOST = "http://rs-as0.qbox.me"
				conf.IO_HOST = "http://iovip-as0.qbox.me"
			}
		}
	}
	m3u8s := qfetch.Fetch(&mac, job, checkExists, file, bucket, worker, logFile)
	//fmt.Println(m3u8s)
	for i := range m3u8s {
		_, err := qfetch.PutFile(m3u8s[i], bucket, strings.TrimPrefix(m3u8s[i], "new_"), accessKey, secretKey)
		if err != nil {
			fmt.Println("Upload new m3u8 file failed, ", err)
		}

		// add check if file exists or not
		_, statErr := os.Stat(m3u8s[i])
		if statErr == nil {
			deleteErr := os.Remove(m3u8s[i])
			for deleteErr != nil {
				fmt.Println("Delete new m3u8 file failed, ", deleteErr)
				deleteErr = os.Remove(m3u8s[i])
			}
		}
	}
}
