package nb

import (
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/kardianos/osext"
)

var nbfeconn net.Conn
var nbfechan chan map[string]interface{}
var doNB bool
var proghash string
var progname string
var hostname string
var nbfedone chan bool

func init() {
	nbfe := os.Getenv("NB_FRONTEND")
	if nbfe == "" {
		fmt.Println("WARN: no $NB_FRONTEND")
		doNB = false
		return
	}
	doNB = true
	nbfedone = make(chan bool)
	filename, err := osext.Executable()
	if err != nil {
		panic(err)
	}
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	hostname = os.Getenv("NB_HOSTNAME")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	progname = os.Args[0]
	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		panic(err)
	}
	proghash = fmt.Sprintf("%x", h.Sum(nil))
	nbfechan = make(chan map[string]interface{}, 1000)
	nbfeconn, err = net.Dial("tcp", nbfe)
	if err != nil {
		panic(err)
	}
	go sender()
}

func sender() {
	if !doNB {
		return
	}
	enc := gob.NewEncoder(nbfeconn)
	for doc := range nbfechan {
		err := enc.Encode(&doc)
		if err != nil {
			panic(err)
		}
	}
	nbfeconn.Close()
	close(nbfedone)
}

func NB(measurement string, params ...interface{}) {
	if !doNB {
		return
	}
	doc := make(map[string]interface{})
	doc["m"] = measurement
	if len(params)%2 != 0 {
		panic("NB params must be even numbered")
	}
	for i := 0; i < len(params); i += 2 {
		doc[params[i].(string)] = params[i+1]
	}
	doc["progname"] = progname
	doc["proghash"] = proghash
	doc["hostname"] = hostname
	_, ok := doc["sourcetime"]
	if !ok {
		doc["sourcetime"] = time.Now().UnixNano()
	}
	select {
	case nbfechan <- doc:
	default:
		nbfechan <- doc
		if measurement != "overflow" {
			NB("nb.overflow")
		}
	}
}

func NBClose() {
	NB("nb.close")
	close(nbfechan)
	<-nbfedone
}
