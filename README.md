# gokit simple workshop

## 01
- build simple web seervice with buildin `net/http` pacage
    ```sh
    # switch 01 foldeer
    cd 01

    # run simple web service
    go run main.go

    # visit web service
    curl localhost:8080
    ```

## 02
- create struct and implement `handler` interface `ServeHTTP(ResponseWriter, *Request)`
    ```sh
    # switch 02 foldeer
    cd 02

    # run simple web service
    go run main.go

    # visit web service
    curl -X POST localhost:8080 -d '{"a":1, "b": 1}'
    ```

## 03
- Extend 02 sample to build 
- Extract Gokit basic component to demo who Gokit does work: `Service`,`Endpoint`,`Transport`
    ```sh
    # switch 03 foldeer
    cd 03

    # run simple web service
    go run main.go

    # visit web service
    curl localhost:8080
    ```

## 04 
- generator gokit basic service by toolchain [cage1016/gk: Go-Kit Genetator](https://github.com/cage1016/gk)
    ```sh
    # switch 04 foldeer
    cd 04

    # run simple web service
    go run cmd/square/main.go

    # visit web service
    curl -X POST -d '{"s": 5}' localhost:8180/square
    ```