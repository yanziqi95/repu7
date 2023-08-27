package main

func (cli *CLI) election(myAddress string) {
	//获取上一个区块的选举人
	bc := NewBlockchain() //打开数据库，读取区块链并构建区块链实例
	defer bc.Db.Close()
	bci := bc.Iterator()
	lastBlock := bci.Next()
	target := lastBlock.Ip

	myIp := string(getIPV4())

	sendElection(myIp, myAddress, target[0])
}
