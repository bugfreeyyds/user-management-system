**this is a user management system.**

#run httpserver
`go run httpserver/*.go`

#run tcpserver
`go run tcpserver/cmd/main.go`

#test login request
`curl -XPOST --data "username=username8&passwd=123456" localhost:8080/api/v1/login`
