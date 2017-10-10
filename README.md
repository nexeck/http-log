# HTTP Log Server for debugging purposes

> It enables you to see the content of any http request you send against this service

## pprof
```sh
echo "GET http://localhost:8080/" | vegeta attack -duration=60s | tee results.bin | vegeta report
go tool pprof http://localhost:8082/debug/pprof/profile
```
