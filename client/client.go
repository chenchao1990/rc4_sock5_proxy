
package main

import (
	"io"
	"crypto/rc4"
	"log"
	"crypto/md5"
	"net"
	"fmt"
	"sync"
	"runtime"
	"errors"
	"strings"
)
func err_line(err error) error {
	_, _, line, _ := runtime.Caller(1)
	s := []string{"line: ", string(line), err.Error()}
	err_str := errors.New(strings.Join(s, " "))
	fmt.Println(err_str)
	return err_str
}
// *********Write*************

type CryptoWriter struct {
	w io.Writer
	cipher *rc4.Cipher
}

func NewCrytoWriter(s io.Writer, key string) io.Writer  {
	MD5sum := md5.Sum([]byte(key))
	cipher_str, err := rc4.NewCipher(MD5sum[:])

	if err != nil {
		log.Println(err)
	}
	return &CryptoWriter{w: s, cipher: cipher_str, }
}

func (w *CryptoWriter) Write(b []byte) (int, error)  {
	buf := make([]byte, len(b))

	w.cipher.XORKeyStream(buf, b)
	w.w.Write(buf)
	return len(buf), nil
}

//   *************Read****************

type CryptoReader struct {
	r io.Reader
	cipher *rc4.Cipher
}

func NewCryReader(s io.Reader, key string) io.Reader {
	MD5sum := md5.Sum([]byte(key))
	cipher_str, err := rc4.NewCipher(MD5sum[:])
	if err != nil{
		log.Fatal(err)
	}
	return &CryptoReader{r:s,cipher:cipher_str}
}

func (r *CryptoReader) Read(p []byte) (n int, err error)  {
	n, err = r.r.Read(p)
	r.cipher.XORKeyStream(p[:n], p[:n])
	return n, err
}


func handleConn(conn net.Conn){
	defer conn.Close()
	// 请求vps 的10086端口
	remote, err := net.Dial("tcp", "67.216.205.28:12306")
	if err != nil{
		err_line(err)
		return
	}
	defer remote.Close()
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// 从本地发送到vps
	go func() {
		defer wg.Done()
		w := NewCrytoWriter(remote, "pwd@1234")
		io.Copy(w, conn)
		remote.Close()
	}()
	// 从vps接收数据 到本地
	go func() {
		defer wg.Done()
		r := NewCryReader(remote, "pwd@1234")
		io.Copy(conn, r)
		remote.Close()
	}()

	wg.Wait()
	return

}

func main() {
	listener, err := net.Listen("tcp", ":8888")	// 监听本地端口 接受请求
	if err !=nil{
		log.Fatal(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()	// 等待本地的请求
		fmt.Println(conn.RemoteAddr())
		if err != nil{
			fmt.Println("Client:::listener accept conn Error: ", err)
			continue
		}
		go handleConn(conn)

	}

}


