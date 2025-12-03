//go:build generate

package main

//go:generate go run ../../cmd/guix/main.go -input=app.gx -output=app_gen.go
//go:generate go run ../../cmd/guix/main.go -input=chart.gx -output=chart_gen.go
