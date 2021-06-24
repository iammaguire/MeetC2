package main

import (
    "io/ioutil"
)
func main() {
	d1 := []byte("BEAasdadasdasdasdCON asdadadasdadsd\n")
	ioutil.WriteFile("C:\\Users\\meet\\Desktop\\BEACONTEST.txt", d1, 0644)
}