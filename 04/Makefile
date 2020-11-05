s ?= 2
square:
	curl -X POST -d '{"s":${s}}' localhost:8180/square

generate:
	go mod init github.com/cage1016/square
	./gk n s square
	sed -i "" 's|Foo(ctx context.Context, s string) (res string, err error)|Square(ctx context.Context, s int64) (res int64, err error)|g' internal/app/square/service/service.go
	./gk init square
	sed -i "" 's/return res, err/return s * s, err/g' internal/app/square/service/service.go
	./gk add grpc square
	cd pb/square && ./compile.sh
	./gk init grpc square
	./gk new cmd square