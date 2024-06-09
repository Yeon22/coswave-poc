package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type Config struct {
    Nodes []*NodeConfig `json:"nodes"`
    Port  int           `json:"port"`
}

type NodeConfig struct {
    ID       int `json:"id"`
    BPMLimit int64 `json:"bpm_limit"`
    RPMLimit int `json:"rpm_limit"`
}

type Node struct {
    NodeConfig
    currentBPM  int64
    currentRPM  int
}

type LoadBalancer struct {
	nodes   []*Node // slice는 배열과 유사하지만 길이를 원하는대로 늘리거나 줄일 수 있음
	mux     sync.Mutex
}

func NewLoadBalancer() *LoadBalancer {
    // & 주소 참조, * 주소가 가리키는 실제
    // make() 함수는 런타임 초기화시 필요한 데이터구조(slice, map, channel)를 초기화할 때 사용
	return &LoadBalancer{nodes: make([]*Node, 0)}
}

func NewNode(nodeConfig *NodeConfig) *Node {
    return &Node{
        NodeConfig: *nodeConfig,
        currentBPM: 0,
        currentRPM: 0,
    }
}

func (lb *LoadBalancer) AddNode(nodeConfig *NodeConfig) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

    node := NewNode(nodeConfig)
	lb.nodes = append(lb.nodes, node)
}

func (lb *LoadBalancer) ProcessRequest(bodySize int64) bool {
	lb.mux.Lock()
	defer lb.mux.Unlock()

    // 모든 노드에 대해 요청 처리를 시도합니다.
    for _, node := range lb.nodes {
        increasedRPM := node.currentRPM + 1
        increasedBPM := node.currentBPM + bodySize

        // 노드의 요청 및 바이트 제한을 확인합니다.
        if increasedRPM <= node.RPMLimit && increasedBPM <= node.BPMLimit {
            node.currentBPM = increasedBPM
            node.currentRPM = increasedRPM

            fmt.Printf("Request processed by node %d\n", node.ID)
            return true // 요청이 처리되었습니다.
        }
    }

	return false
}

func main() {
	config, err := readConfig("config.json")
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	lb := NewLoadBalancer()
	for _, node := range config.Nodes {
		lb.AddNode(node)
	}

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println(r.ContentLength)
        if lb.ProcessRequest(r.ContentLength) {
            fmt.Printf("Request processed successfully\n")
        } else {
            fmt.Printf("All nodes are overloaded. Request rejected\n")
            http.Error(w, "All nodes are overloaded. Request rejected", http.StatusServiceUnavailable)
            return
        }

        // 모든 요청 처리가 끝나면 "Done" 응답 보내기
        // 실제 프로젝트에선 역방향 프록시를 호출하거나 할 듯
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Done"))
    })

	fmt.Println("Server listening on port: ", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}

func readConfig(filename string) (*Config, error) {
	var config *Config

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
