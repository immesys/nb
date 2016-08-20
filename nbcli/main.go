package main

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/immesys/nb"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: nb <measurement> <key> <value> <key> <value> ...")
		fmt.Println(" env vars: NB_HOSTNAME, NB_FRONTEND")
		fmt.Println(" if <value> is @file the next argument will be treated as a filename")
		fmt.Println(" if <value> is @stdin the stdin will be associated with this value")
		os.Exit(1)
	}
	if os.Getenv("NB_FRONTEND") == "" {
		fmt.Println("you need $NB_FRONTEND")
		os.Exit(1)
	}
	params := []interface{}{}
	measurement := os.Args[1]
	donestdin := false
	for i := 2; i < len(os.Args); i++ {
		params = append(params, os.Args[i]) //key
		i++
		switch os.Args[i] {
		case "@file":
			i++
			fname := os.Args[i]
			f, err := os.Open(fname)
			if err != nil {
				fmt.Println("file open: ", err)
				os.Exit(1)
			}
			v, err := ioutil.ReadAll(f)
			if err != nil {
				fmt.Println("file read: ", err)
				os.Exit(1)
			}
			params = append(params, string(v))
		case "@stdin":
			if donestdin {
				fmt.Println("can only do one @stdin")
				os.Exit(1)
			}
			donestdin = true
			v, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println("stdin read: ", err)
				os.Exit(1)
			}
			params = append(params, string(v))
		default:
			params = append(params, os.Args[i])
		}
	}
	NB(measurement, params...)
	NBClose()
}
