package main

import (
	"io"
	"crypto/rc4"
	"crypto/md5"
	"fmt"

	"net"
)


// *********Write*************

type CryptoWriter struct {
	w io.Writer
	cipher *rc4.Cipher
}

func NewCrytoWriter(s io.Writer, key string) io.Writer  {

	MD5sum := md5.Sum([]byte(key))
	cipher_str, err := rc4.NewCipher(MD5sum[:])
	if err != nil {
		fmt.Println("ERROR: new Writer:::", err)
	}
	return &CryptoWriter{w: s, cipher: cipher_str, }
}

func (w *CryptoWriter) Write(b []byte) (int, error)  {
	buf := make([]byte, len(b))

	w.cipher.XORKeyStream(buf, b)
	w.w.Write(buf)
	return len(buf), nil
}

// *********conn Write*************

type ConntoWriter struct {
	w net.Conn
	cipher *rc4.Cipher
}

func NewConnWriter(s net.Conn, key string) io.Writer  {

	MD5sum := md5.Sum([]byte(key))
	cipher_str, err := rc4.NewCipher(MD5sum[:])
	if err != nil {
		fmt.Println("ERROR: new Writer:::", err)
	}
	return &ConntoWriter{w: s, cipher: cipher_str}
}

func (w *ConntoWriter) Write(b []byte) (int, error)  {
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
		fmt.Println("ERROR: new Reader:::", err)
	}
	return &CryptoReader{r:s,cipher:cipher_str}
}

func (r *CryptoReader) Read(p []byte) (n int, err error)  {

	n, err = r.r.Read(p)
	if err != nil{
		fmt.Println("\033[;33m ERROR: interface Read:::\033[0m", err)
		return n, err
	}
	r.cipher.XORKeyStream(p[:n], p[:n])
	return n, nil
}
//   ************* conn Read****************

type ConntoReader struct {
	r io.Reader
	cipher *rc4.Cipher
}

func NewConnReader(s net.Conn, key string) io.Reader {
	MD5sum := md5.Sum([]byte(key))
	cipher_str, err := rc4.NewCipher(MD5sum[:])
	if err != nil{
		fmt.Println("ERROR: new Reader:::", err)
	}
	return &CryptoReader{r:s,cipher:cipher_str}
}

func (r *ConntoReader) Read(p []byte) (n int, err error)  {
	n, err = r.r.Read(p)
	if err != nil{
		fmt.Println("\033[;33m ERROR: interface Read:::\033[0m", err)
		return n, err
	}
	r.cipher.XORKeyStream(p[:n], p[:n])
	return n, nil
}
