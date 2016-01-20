#oci2docker - Convert OCI bundle to docker image

oci2docker is a small library and CLI binary that converts [OCI](https://github.com/opencontainers/specs) bundle to docker image. It takes OCI bundle as input, and gets docker image as output.

## How it works

An important work of oci2docker is converting OCI Specs to Dockerfile, below is a list of conversion principle:

### config.json

#### Root Configuration
|OCI Specs|Dockerfile|
|---------|----------|
| path | ADD |

#### Process configuration
|OCI Specs|Dockerfile|
|---------|----------|
| env | ENV |
| cwd | WORKDIR |
| args | ENTRYPOINT |
| user | USER |

#### Mount Points
|OCI Specs|Dockerfile|
|---------|----------|
| mounts | VOLUME |

### runtime.json

#### Mount Configuration
|OCI Specs|Dockerfile|
|---------|----------|
| mounts | VOLUME |


## Build

Installation is as simple as:

```bash
go get github.com/huawei-openlab/oci2docker
```

or as involved as:

```bash
# create a 'github.com/huawei-openlab' in your GOPATH/src
cd $GOPATH/src/github.com/
mkdir huawei-openlab
cd huawei-openlab
git clone https://github.com/huawei-openlab/oci2docker.git
cd oci2docker
go build
```

## Usage

```
$ ./oci2docker
NAME:
   oci2docker - A tool for coverting oci bundle to docker image

USAGE:
   oci2docker [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   convert      convert operation
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h           show help
   --version, -v        print the version

$ ./oci2docker convert
NAME:
   convert - convert operation

USAGE:
   command convert [command options] [arguments...]

OPTIONS:
   --oci-bundle "oci-bundle"    path of oci-bundle to convert
   --image-name "image-name"    docker image name
   --port                       exposed port of docker images
```

## Example

```
$ ./oci2docker convert --oci-bundle example/oci-bundle/ --image-name cts/hello-docker
DEBU[0000] Docker build context is in /tmp/oci2docker911067878
 
Sending build context to Docker daemon 1.031 MB
Sending build context to Docker daemon 
Step 0 : FROM scratch
 ---> 
Step 1 : MAINTAINER ChengTiesheng <chengtiesheng@huawei.com>
 ---> Using cache
 ---> c6fd471de5d0
Step 2 : ADD ./rootfs .
 ---> 319ffc8c52a7
Removing intermediate container 4bdbe7abd980
Step 3 : ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin TERM xterm
 ---> Running in 5fde4642d6d3
 ---> 11ec7f4bdb2f
Removing intermediate container 5fde4642d6d3
Step 4 : ENTRYPOINT /bin/sh
 ---> Running in d47b25f16d80
 ---> 2f6dedb309af
Removing intermediate container d47b25f16d80
Successfully built 2f6dedb309af

$ docker run -i -t cts/hello-docker
/ # 

```
