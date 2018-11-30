package qfetch

import (
	"fmt"

	"github.com/qiniu/api.v6/auth/digest"
	"github.com/qiniu/api.v6/rs"
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
