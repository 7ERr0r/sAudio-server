
# Building
go get github.com/southskies/golang-opus
go get go build github.com/Szperak/sAudio-server/server
go build github.com/Szperak/sAudio-server/server/cmd/mixserver

# Running (on default port)
./mixserver
