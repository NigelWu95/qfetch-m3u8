package qfetch

import (
	"context"
	"fmt"

	"github.com/qiniu/api.v6/auth/digest"
	"github.com/qiniu/api.v6/rs"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/rpc"
)

// BucketInfo to bucket info structure
type BucketInfo struct {
	Region string `json:"region"`
}

// domain of quering bucket
var (
	BucketRsHost = "http://rs.qiniu.com"
)

/*
GetBucketInfo to get bucket info
@param mac
@param bucket - bucket name
@return bucketInfo, err
*/
func GetBucketInfo(mac *digest.Mac, bucket string) (bucketInfo BucketInfo, err error) {
	client := rs.New(mac)
	bucketURI := fmt.Sprintf("%s/bucket/%s", BucketRsHost, bucket)
	callErr := client.Conn.Call(nil, &bucketInfo, bucketURI)
	if callErr != nil {
		if v, ok := callErr.(*rpc.ErrorInfo); ok {
			err = fmt.Errorf("code: %d, %s, xreqid: %s", v.Code, v.Err, v.Reqid)
		} else {
			err = callErr
		}
	}
	return
}

func putFile(localFile string, bucket string, key string, accessKey, secretKey string) {
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuadong
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)
}
