package main

import (
	"net"
	"bufio"
	"log"
	"io"
	"encoding/binary"
	"fmt"
	"sync"
	"runtime"
	"errors"
	"strings"
	"strconv"
)
// 握手

var key_str = "pwd@1234"

func err_line(err error) error {
	_, _, line, _ := runtime.Caller(1)
	s := []string{"line: ", strconv.Itoa(line), err.Error()}
	err_str := errors.New(strings.Join(s, " "))
	fmt.Println(err_str)
	return err_str
}

func handshake(r *bufio.Reader, cli_wirter io.Writer) error {
	_, err := r.ReadByte()		// 第一个字节 位版本号
	if err != nil {
		return err_line(err)
	}

	n_method, err := r.ReadByte()	// 读取第二个字节 客户端请求类型 1为代理
	if err != nil{
		return err_line(err)
	}

	buf := make([]byte, n_method)		// 读取第三个字节为 客户端支持的验证方式
	io.ReadFull(r, buf)
	resp := []byte{5, 0}		// 收到客户端验证后 回应客户端 0 不需要验证
	cli_wirter.Write(resp)			// 服务端需要客户端提供哪种验证方式信息
	return nil
}

// 获取客户端请求的地址和端口
func readAddr(r *bufio.Reader) (string, error)  {
	_, err:= r.ReadByte()		// 读取第一个字节	位版本号
	if err != nil{
		return "", err_line(err)
	}

	cmd, err := r.ReadByte()		// 读取第二个字节 客户端请求类型 1为代理
	if err != nil{
		return "", err_line(err)
	}
	log.Printf("CMD: %d", cmd)

	r.ReadByte()			// 第三个字节为保留字 RSV

	_, err = r.ReadByte()		//第四个字节为ATYP 请求的远程服务器地址类型 ip domainname ipv6
	if err != nil{
		return "", err_line(err)
	}

	addr_len, err := r.ReadByte()		// 这个字节代表 远程服务器地址的长度
	if err != nil{
		return "", err_line(err)
	}

	addr := make([]byte, addr_len)		// 根据服务器地址的长度  去读取地址
	io.ReadFull(r, addr)

	var port int16
	binary.Read(r, binary.BigEndian, &port)
	return fmt.Sprintf("%s:%d", addr, port), nil
}

// 开始代理
func startProxy(addr string, client *bufio.Reader, new_wirter io.Writer) error {

	remote, err := net.Dial("tcp", addr)

	if err != nil{
		return err_line(err)
	}
	defer remote.Close()
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		// 将请求网站的数据 加密后返回给客户端
		io.Copy(new_wirter, remote)
		//client.Close()
	}()
	go func() {
		// 从客户端读取数据 发给请求的网站
		defer wg.Done()
		io.Copy(remote, client)
		//remote.Close()
	}()
	wg.Wait()
	return nil
}
func handleConn(conn net.Conn){
	defer conn.Close()
	// 将conn先实例化为两个类型
	// 这里的关键在于 之后所有的数据操作 都要基于这两个类型
	// 如果中间有新的实例化 数据传输则会出现错误  两天时间发现的血的教训！！！
	new_conn := NewCryReader(conn, key_str)
	new_writer := NewCrytoWriter(conn, key_str)
	r := bufio.NewReader(new_conn)

	err := handshake(r, new_writer)		// 握手
	if err != nil{
		return
	}
	addr, err := readAddr(r)
	if err != nil{
		fmt.Println("read client Addr err:::", err)
		return
	}
	fmt.Printf("\033[1;32m remote addr is :::::%s \n\033[0m", addr)
	resp := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, err = new_writer.Write(resp)
	if err != nil{
		err_line(err)
		return
	}
	// 开始代理
	err = startProxy(addr, r, new_writer)
	if err != nil{
		return
	}

}

func main() {
	l ,err := net.Listen("tcp", ":12306")
	if err != nil{
		_, _, line, _ := runtime.Caller(0)
		log.Fatal("line: ", line, err)
	}
	defer l.Close()

	for{
		conn, err := l.Accept()
		if err != nil{
			_, _, line, _ := runtime.Caller(0)
			log.Fatal("line",  line, err)
		}

		go handleConn(conn)
	}
}
