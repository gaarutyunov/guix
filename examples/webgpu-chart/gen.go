//go:build generate

package main

//go:generate go run ../../cmd/guix/main.go -input=app.gx -output=app_gen.go
