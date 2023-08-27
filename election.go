package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

type WeightedData struct {
	Data   string
	Weight int
}

type ElectionRequest struct {
	Ip      string
	Address string
}

func sendElection(myIp string, myAddress string, targetIp string) {
	//发送选举信息
	conn, err := net.Dial("tcp", targetIp+":9885")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	// 准备选举请求数据，使用结构体
	request := ElectionRequest{
		Ip:      myIp,
		Address: myAddress,
	}

	// 将选举请求结构体编码为JSON格式
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error encoding request:", err.Error())
		return
	}

	// 发送选举请求给服务端
	_, err = conn.Write(requestJSON)
	if err != nil {
		fmt.Println("Error sending data:", err.Error())
		return
	}

	// 等待服务端响应
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		fmt.Println("Error reading response:", err.Error())
		return
	}

	// 处理服务端响应
	fmt.Println("Server response:", string(response))
}

var electionList = []WeightedData{}

func recvElection(nonce int) []string {
	//处理选举信息
	//清空选举列表
	electionList = []WeightedData{}
	// 监听端口
	listener, err := net.Listen("tcp", ":9885")
	if err != nil {
		fmt.Println("Error listening:", err.Error())

	}
	defer listener.Close()
	fmt.Println("Server is listening on 9885 to receive election request")

	//30 seconds for accept elction request
	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			// 超时后关闭监听器
			fmt.Println("30-second timeout reached, closing the server.")

		default:
			// 等待客户端连接
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting:", err.Error())
				continue
			}

			// 处理客户端请求
			go handleElection(conn)
		}
	}
	fmt.Println("完成收集选举列表", electionList)
	electors := random(int64(nonce), electionList, selectSize)
	fmt.Println(electors)
	return electors
}

func handleElection(conn net.Conn) {
	// 读取客户端发送的JSON数据
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	// 解码JSON数据为ElectionRequest结构体
	var request ElectionRequest
	err = json.Unmarshal(buffer[:n], &request)
	if err != nil {
		fmt.Println("Error decoding JSON:", err.Error())
		return
	}

	//加入列表
	weight := getElectorBalance(request.Address) + 1

	// 创建WeightedData结构体并添加到electionList
	data := WeightedData{
		Data:   request.Ip,
		Weight: weight,
	}

	electionList = append(electionList, data)
	// 处理选举请求，这里可以根据需要实现选举逻辑
	// 访问request.IP和request.Address字段以获取客户端的IP地址和其他相关信息

	// 发送响应给客户端
	response := "Election request added to list."
	conn.Write([]byte(response))
	conn.Close()
}

func random(randomSeed int64, data []WeightedData, n int) []string {
	//随机数发生器

	rand.Seed(randomSeed)
	totalWeight := 0
	for _, data := range data {
		totalWeight += data.Weight
	}

	// 选择n个数据
	selectedData := map[string]bool{} // 用于跟踪已选中的数据

	for len(selectedData) < n {
		// 生成一个随机数，范围在0到总权重之间
		randomNumber := rand.Intn(totalWeight)

		// 根据随机数选择数据
		for _, data := range data {
			if randomNumber < data.Weight && !selectedData[data.Data] {
				selectedData[data.Data] = true
				fmt.Printf("选择的数据: %s\n", data.Data)
				break
			}
			randomNumber -= data.Weight
		}
	}
	//返回成功的选举人的ip
	electors := []string{}
	for elector := range selectedData {
		electors = append(electors, elector)
	}
	return electors
}

func getElectorBalance(address string) int {
	if !ValidateAddress(address) {
		log.Panic("ERROR: 地址非法")
	}
	bc := NewBlockchain()
	defer bc.Db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("'%s'的账号余额是: %d\n", address, balance)
	return balance
}
