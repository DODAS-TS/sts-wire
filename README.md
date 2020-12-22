[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/DODAS-TS/sts-wire)

# STS WIRE 

## Requirements

- fuse installed
## Quick start

Download the binary from the latest release on [github](https://github.com/DODAS-TS/dodas-go-client/releases). For instance:

```bash
wget https://github.com/DODAS-TS/sts-wire/releases/download/v0.0.10/sts-wire
chmod +x sts-wire
cp sts-wire /usr/local/bin
```
## Building from source

To compile from the sources you need a `Go` version that supports `Go modules` (e.g. >= v1.12). You can compile the executable using the `Makefile`:

```bash
# Linux
make build
# Windows
make build-windows
# MacOS
make build-macos
```

## How to use

You can see how to use the program asking for help in the command line:

```bash
./sts-wire -h
```

The result of the above command will be something similar to this:

```text
Usage:
  sts-wire <instance name> <s3 endpoint> <rclone remote path> <local mount point> [flags]

Flags:
      --config string           config file (default "./config.json")
  -h, --help                    help for sts-wire
      --insecureConnection      check the http connection certificate (default true)
      --log string              where the log has to write, a file path or stderr (default "stderr")
      --noPassword              to not encrypt the data with a password
      --refreshTokenRenew int   time span to renew the refresh token in minutes (default 10)
```

As you can see, to use the `sts-wire` you need the following arguments to be passed:

- `<instance name>`: the name you give to the `sts-wire` instance
- `<s3 endpoint>`: the *s3* server you want to use, also a local one, e.g. `http://localhost:9000`
- `<rclone remote path>`: the remote path that you need to mount locally, relative to the *s3* server, e.g. `/folder/on/my/s3`. It could be any of your buckets, also root `/`.
- `<local mount point>`: the folder where you want to mount the remote source. It could be also relative to the current working folder, e.g. `./my_local_mountpoint`
  
**Note:** remember to set the `IAM_SERVER` variable before using the command, to point to a trusted IAM server.

```bash
export IAM_SERVER="https://my.iam.server.com"
```

Alternatively, you can create a YAML configuration file like the following:

```yaml
---
instance_name: test_instance
s3_endpoint: http://localhost:9000
rclone_remote_path: /test
local_mount_point: ./my_local_mountpoint
IAM_Server: https://my.iam.server.com
```

### Launch the program

In the following example you can see how the program is launched:

```bash
IAM_SERVER="https://my.iam.server.com" ./sts-wire myMinio https://myserver.com:9000 / ./mountedVolume
# Using a config file name myConfig.yml
./sts-wire --config myConfig.yml
```


After that, you have to follow all the instructions and providing a password for credentials encryption when requested.
Eventually, if everything went well, on your browser you will be prompted with a message like:

```text
VOLUME MOUNTED, YOU CAN NOW CLOSE THIS TAB. 
```

The volume will stay mounted untill you exit the running sts-wire process with Control-C
## Contributing

If you want to contribute:

1. create a branch
2. upload your changes
3. create a pull request

Thanks!
