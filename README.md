# fssgo
basic video service for banyan platform

### 编译环境

- 编译 Linux/macOS 版本 (在 bash 环境下执行) Build linux/macOS version

        创建目录/home/go/
        下载go1.11.4.linux-amd64.tar.gz到/home/go 目录
        tar zxf go1.11.4.linux-amd64.tar.gz
        mv go go1.11
        export PATH=$PATH:/home/go/go1.11/bin

        
### 编译命令

- 获取代码

        创建目录/home/go/bvs
        mkdir -p /home/go/bvs
        export GOPATH=/home/go/bvs
        
        创建目录$GOPATH/src/golang.org/x
        mkdir -p $GOPATH/src/golang.org/x
        cd $GOPATH/src/golang.org/x
        git clone https://github.com/golang/sys.git
        
        创建目录$GOPATH/src/github.com
        mkdir -p $GOPATH/src/github.com
        cd $GOPATH/src/github.com
        
        下载bvs代码
        git clone https://github.com/banyanteam/bvs.git
        
        获取其余依赖库
        go get -u github.com/gin-gonic/gin
        go get -u github.com/kardianos/govendor
        
        将govendor添加到PATH
        export PATH=$GOPATH/bin:$PATH
        
        cd bvs
        go build
