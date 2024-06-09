package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Node struct {
	ID       int `json:"id"`
	BPMLimit int `json:"bpm_limit"`
	RPMLimit int `json:"rpm_limit"`
}

type LoadBalancer struct {
	nodes      []*Node // slice는 배열과 유사하지만 길이를 원하는대로 늘리거나 줄일 수 있음
	requests   map[int]int
	bytes      map[int]int
	mux sync.Mutex
}

func NewLoadBalancer() *LoadBalancer {
    // & 주소 참조, * 주소가 가리키는 실제
	return &LoadBalancer{
		nodes:    make([]*Node, 0), // make() 함수는 런타임 초기화시 필요한 데이터구조(slice, map, channel)를 초기화할 때 사용
		requests: make(map[int]int), // map[KeyType]ValueType
		bytes:    make(map[int]int),
	}
}

func (lb *LoadBalancer) AddNode(node *Node) {
	lb.mux.Lock()
	defer lb.mux.Unlock()
	lb.nodes = append(lb.nodes, node)
}

func (lb *LoadBalancer) ProcessRequest(nodeID int, bodySize int) bool {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	lb.requests[nodeID]++
	lb.bytes[nodeID] += bodySize

	node := lb.getNodeByID(nodeID)
	if node == nil {
		return false
	}

	if lb.requests[nodeID] > node.RPMLimit || lb.bytes[nodeID] > node.BPMLimit {
		return false
	}

	return true
}

func (lb *LoadBalancer) getNodeByID(nodeID int) *Node {
	for _, node := range lb.nodes {
		if node.ID == nodeID {
			return node
		}
	}
	return nil
}

func main() {
	nodes, err := readConfig("config.json")
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	lb := NewLoadBalancer()
	for _, node := range nodes {
		lb.AddNode(node)
	}

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // request 시뮬레이션
        for i := 0; i < 200; i++ {
            bodySize := 100
            nodeID := 1
            if i%2 == 0 {
                nodeID = 2
            }
            if lb.ProcessRequest(nodeID, bodySize) {
                fmt.Printf("Request %d processed by node %d\n", i+1, nodeID)
            } else {
                fmt.Printf("Request %d rejected by node %d due to rate limit\n", i+1, nodeID)
            }
    
            time.Sleep(time.Millisecond * 100) // 요청 딜레이 시뮬레이션
        }

        // 모든 요청 처리가 끝나면 "Done" 응답 보내기
        // 실제 프로젝트에선 역방향 프록시를 호출하거나 할 듯
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Done"))
    })

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func readConfig(filename string) ([]*Node, error) {
	var nodes []*Node

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
