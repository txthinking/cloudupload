# Cloud Upload

Upload files to multiple cloud storage in parallel.

### Download

#### Binary

Download from (releases)[https://github.com/txthinking/cloudupload/releases] page.

#### Source

```
$ go get github.com/txthinking/cloudupload/cli/cloudupload
```

### Run

```
NAME:
   Cloud Upload - Upload files to multiple cloud storage in parallel

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   20181008

AUTHOR:
   Cloud <cloud@txthinking.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d                    Enable debug, more logs
   --debugListen value, -l value  Listen address for debug (default: "127.0.0.1:6060")
   --listen value                 Listen address
   --domain value                 If domain is specified, 80 and 443 ports will be used. Listen address is no longer needed
   --maxBodySize value            Max size of http body, M (default: 0)
   --timeout value                Read timeout, write timeout x2, idle timeout x20, s (default: 0)
   --origin value                 Allow origins for CORS, can repeat more times. like https://google.com, suggest add https://google.com/ too
   --enableLocal                  Enable local store
   --localStorage value           Local directory path
   --enableGoogle                 Enable google store, first needs $ gcloud auth application-default login
   --googleBucket value           Google bucket name
   --enableAliyun                 Enable aliyun OSS
   --aliyunAccessKeyID value      Aliyun access key id
   --aliyunAccessKeySecret value  Aliyun access key secret
   --aliyunEndpoint value         Aliyun endpoint, like: https://oss-cn-shanghai.aliyuncs.com
   --aliyunBucket value           Aliyun bucket name
   --enableTencent                Enable Tencent
   --tencentSecretId value        Tencent secret id
   --tencentSecretKey value       Tencent secret key
   --tencentHost value            domain
   --help, -h                     show help
   --version, -v                  print the version
```

### Upload

#### Request

* Method: `POST`
* Header:
    * `Accept`, one of:
        * `application/json`
        * `text/plain`
    * `Content-Type`, one of:
        * `application/octet-stream`
        * `application/base64`
        * `multipart/form-data...`, field name: `file`
    * `X-File-Name`:
        * Full file name with suffix, Only required when `Content-Type` is `application/octet-stream` or `application/base64`.
* Body, one of:
    * Binary file content when `Content-Type` is `application/octet-stream`.
    * Base64 encoded file content when `Content-Type` is `application/octet-stream`.
    * Multipart form data when `Content-Type` is `multipart/form-data...`.

#### Response

* Status Code: 200
    * Content-Type, one of:
        * `application/json`
        * `text/plain; charset=utf-8`
    * Body:
        * File path
        * `{ "file": "file path" }`
* Status Code: !200
    * Content-Type:
        * `text/plain; charset=utf-8`
    * Body:
        * Error message

### Example

```
$ curl -H 'Content-Type: application/octet-stream' -H 'X-File-Name: Angry.png' --data-binary @Angry.png https://yourdomain.com
vbpovzsdzbxu/Angry.png
$ curl -F 'file=@Angry.png' https://yourdomain.com
vbpovzsdzbxu/Angry.png
```
