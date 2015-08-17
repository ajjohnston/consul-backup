package main

import (
	"fmt"
    "sort"
    "os"
    "io/ioutil"
    "encoding/json"
    "github.com/docopt/docopt-go"
    consul "github.com/hashicorp/consul/api"
)


//type KVPair struct {
//    Key         string
//    CreateIndex uint64
//    ModifyIndex uint64
//    LockIndex   uint64
//    Flags       uint64
//    Value       []byte
//    Session     string
//}

type ByCreateIndex consul.KVPairs

func (a ByCreateIndex) Len() int           { return len(a) }
func (a ByCreateIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
//Sort the KVs by createIndex
func (a ByCreateIndex) Less(i, j int) bool { return a[i].CreateIndex < a[j].CreateIndex }


func backup(ipaddress string, outfile string) {

    config := consul.DefaultConfig()
    config.Address = ipaddress


	client, _ := consul.NewClient(config)
	kv := client.KV()

	pairs, _, err := kv.List("/", nil)
    if err != nil {
            panic(err)
        }
    sort.Sort(ByCreateIndex(pairs))

    outpairs := map[string]string{}
	for _, element := range pairs {
        outpairs[element.Key] += string(element.Value)
	}

    data, err := json.MarshalIndent(outpairs, "", " ")
    if err != nil {
        panic(err)
    }

    file, err := os.Create(outfile)
    if err != nil {
        panic(err)
    }

    if _, err := file.Write([]byte(data)[:]); err != nil {
        panic(err)
    }
}

func restore(ipaddress string, infile string) {

    config := consul.DefaultConfig()
    config.Address = ipaddress

    data, err := ioutil.ReadFile(infile)
    if err != nil {
        panic(err)
    }

    inpairs := map[string]string{}
    err = json.Unmarshal(data, &inpairs)
    if err != nil {
        panic(err)
    }

    client, _ := consul.NewClient(config)
    kv := client.KV()

    for k, v := range inpairs {
        fmt.Printf("restoring %s:%s\n", k, v)
        p := &consul.KVPair{Key: k, Value: []byte(v)}
        _, err := kv.Put(p, nil)
        if err != nil {
            panic(err);
        }
    }
}

func main() {

    usage := `Consul Backup and Restore tool.

Usage:
  consul-backup [-i IP:PORT] [--restore] <filename>
  consul-backup -h | --help
  consul-backup --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  -i, --address=IP:PORT  The HTTP endpoint of Consul [default: 127.0.0.1:8500].
  -r, --restore     Activate restore mode`

    arguments, _ := docopt.Parse(usage, nil, true, "consul-backup 1.0", false)
    // fmt.Println(arguments)
    if arguments["--restore"] == true {
		fmt.Println("Restore mode:")
		fmt.Printf("Warning! This will overwrite existing kv. Press [enter] to continue; CTL-C to exit")
		fmt.Scanln()
		fmt.Println("Restoring KV from file: ", arguments["<filename>"].(string))
        restore(arguments["--address"].(string), arguments["<filename>"].(string))
    } else {
		fmt.Println("Backup mode:")
		fmt.Println("KV store will be backed up to file: ", arguments["<filename>"].(string))
        backup(arguments["--address"].(string), arguments["<filename>"].(string))
    }

}
