package main

import (
    "io/ioutil"
)
func main() {
	d1 := []byte("test\n")
	ioutil.WriteFile("C:\\Users\\meet\\Desktop\\s.txt", d1, 0644)
}