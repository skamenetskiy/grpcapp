package main

import (
	"sample/internal/service"
	"sample/pkg"

	"github.com/skamenetskiy/grpcapp"
)

func main() {
	grpcapp.Listen(
		grpcapp.WithServiceImplementation(&pkg.Greeter_ServiceDesc, new(service.Service)),
	)
}
