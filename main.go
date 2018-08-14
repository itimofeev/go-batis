package main

import (
	_ "github.com/lib/pq"
	"log"
)

func main() {

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
