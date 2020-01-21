$env:CGO_ENABLED=0
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build .