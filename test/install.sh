PROJECT_PATH=$(pwd)

installPackages(){
    echo "start install dependencies..."
    go get -u github.com/jinzhu/gorm
    go get -u github.com/beego/beego/v2
    go get -u github.com/gin-gonic/gin
    go get -u github.com/go-redis/redis
    go get -u google.golang.org/grpc
    echo "install packages successfully."
}

setupDB() {
    echo "setup db..."
    cd $PROJECT_PATH
    go build ./test/initdb.go
    ./initdb -c conf/tcpserver.yaml
    echo "setup db successfully."
}

installPackages
setupDB
