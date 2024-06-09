# 콕스웨이브 PoC 과제

Rate Limit 기능을 가진 여러 노드로 실행할 수 있는 로드 밸런서 구현

1. 각 노드는 별도의 Rate Limit 을 가질 수 있어야 함
2. BPM(http body Bytes Per Minute), RPM(Requests Per Minute) 두 방법으로 Rate Limit 을 측정
3. 먼저 발생하는 상황에 따라 어떤 옵션이든 실행 될 수 있어야 함

### 프로젝트 실행

```
go run main.go
```

### request 시뮬레이션

```
curl -i "localhost:8080"
```
