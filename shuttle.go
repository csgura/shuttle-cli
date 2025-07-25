package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/csgura/fp"
	"github.com/csgura/fp/option"
	"github.com/csgura/fp/ord"
	"github.com/csgura/fp/slice"
)

func log(fmtstr string, args ...any) {
	fmt.Println(fmt.Sprintf(fmtstr, args...))
}

type Host struct {
	Parent []string
	Name   string
	Alias  string
	Cmd    string
	Theme  string
}

func get(v map[string]any, k string) fp.Option[string] {
	if ret, ok := v[k]; ok {
		if rv, ok := ret.(string); ok {
			return option.Some(rv)
		}
	}
	return option.None[string]()
}

func getList(group []string, v any) []Host {
	switch t := v.(type) {
	case nil:
		return nil
	case map[string]any:
		if t["name"] != nil {
			return slice.Of(Host{
				Parent: group,
				Name:   get(t, "name").OrZero(),
				Cmd:    get(t, "cmd").OrZero(),
				Alias:  get(t, "alias").OrZero(),
				Theme:  get(t, "theme").OrZero(),
			})
		} else {
			return slice.FlatMap(slice.FromMap(t), func(v fp.Entry[any]) fp.Slice[Host] {
				return getList(append(group, v.I1), v.I2)
			})
		}
	case []any:
		return slice.FlatMap(t, func(v any) fp.Slice[Host] {
			return getList(group, v)
		})

	}
	return nil
}

func getComment(lines []string, k string) fp.Option[string] {

	cfgline := slice.Find(lines, fp.Test(strings.Contains, k))
	if cfgline.IsDefined() {
		tarr := strings.Split(cfgline.Get(), "=")
		for t := range slice.Last(tarr).All() {
			return option.Some(strings.TrimSpace(t))
		}
	}

	return option.None[string]()
}

func sshConfigToHost(lines []string) fp.Option[Host] {
	none := option.None[Host]()
	if len(lines) > 0 {

		f := strings.Fields(lines[0])
		if slice.Get(f, 0) == option.Some("Host") {
			for host := range slice.Get(f, 1).All() {
				ret := Host{
					Name:   host,
					Cmd:    "ssh " + host,
					Theme:  getComment(lines, "shuttle.theme").OrZero(),
					Alias:  getComment(lines, "shuttle.alias").OrZero(),
					Parent: slice.Init(strings.Split(getComment(lines, "shuttle.name").OrZero(), "/")),
				}
				return option.Some(ret)
			}
		}
	}
	return none

}

func connect(c Host) {
	if c.Theme != "" {
		log("set theme to %s.", c.Theme)
		cmdhttp1 := exec.Command("osascript", "-e", fmt.Sprintf(`
				tell application "Terminal"
					set myterm to (first window)
					set current settings of myterm to settings set "%s"
				end tell
				`, c.Theme))
		_, err := cmdhttp1.Output()
		if err != nil {
			log("cmd error : %s", err)
			os.Exit(1)
		}

		// f := strings.Fields(c.Cmd)
		// cmd := slice.Head(f)

	}
	err := syscall.Exec("/bin/bash", slice.Of("-i", "-l", "-c", c.Cmd), os.Environ())
	if err != nil {
		log("exec error : %s", err)
		os.Exit(1)
	}
}
func main() {
	list := func() []Host {
		sj, err := os.ReadFile(os.Getenv("HOME") + "/.shuttle.json")
		if err == nil {

			var cfgmap map[string]any
			err = json.Unmarshal(sj, &cfgmap)
			if err != nil {
				log("shuttle json parse error : %s", err)
				os.Exit(1)
			}

			return getList(nil, cfgmap["hosts"])
		}
		return nil
	}()

	sconfig, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/config")
	lines := strings.Split(string(sconfig), "\n")

	if err == nil {

		splited := slice.Fold(lines, slice.Of([]string{}), func(b [][]string, a string) [][]string {
			s := strings.TrimSpace(a)
			f := strings.Fields(s)
			if slice.Get(f, 0) == option.Some("Host") {
				return append(b, slice.Of(s))
			}
			b[len(b)-1] = append(b[len(b)-1], s)
			return b
		})
		list = append(list, slice.FilterMap(splited, sshConfigToHost)...)

	} else {
		log("open ssh config error : %s", err)
	}

	// for _, v := range list {
	// 	log("group = %s, name = %s, cmd = %s , theme = %s", slice.MakeString(v.Parent, "/"), v.Name, v.Cmd, v.Theme)
	// }

	if len(os.Args) == 1 {
		reader := bufio.NewReader(os.Stdin)
		level := 0
		pmap := map[int][]Host{
			0: list,
		}
		gname := ""
		for {

			grouped := slice.GroupBy(pmap[level], func(a Host) string {
				return slice.Get(a.Parent, level).OrZero()
			})

			leafs := slice.Sort(grouped[""], ord.FromCompare(func(a, b Host) int {
				return ord.Given[string]().Compare(a.Name, b.Name)
			}))

			delete(grouped, "")
			slist := slice.Sort(slice.FromMapKeys(grouped), ord.Given[string]())

			if len(leafs) > 0 {
				log("")
				if level > 0 {
					log("Hosts in %s:", gname)
				} else {
					log("Hosts:")

				}
				for i := 0; i < len(leafs); i++ {
					if leafs[i].Alias != "" {
						log("  %d: %s (%s)", i, leafs[i].Name, leafs[i].Alias)

					} else {
						log("  %d: %s", i, leafs[i].Name)
					}
				}
			}

			if len(slist) > 0 {
				log("")
				log("Groups:")
				for i := 0; i < len(slist); i++ {
					log("  %d: %s", i+len(leafs), slist[i])
				}
			}

			if level > 0 {
				log("  up: goto upper level")
			}

			log("")
			fmt.Printf("enter number: ")
			str, _ := reader.ReadString('\n')
			str = strings.TrimSpace(str)
			if str == "up" {
				level = level - 1
				gname = slice.MakeString(slice.Init(strings.Split(gname, "/")), "/")
				continue
			}
			n, err := fp.ParseInt(str).Unapply()
			if err == nil {
				if n < len(leafs) {
					connect(leafs[n])
					os.Exit(0)
				}

				n = n - len(leafs)
				if n < len(slist) {
					gname = gname + "/" + slist[n]
					level = level + 1
					pmap[level] = grouped[slist[n]]
				}
			}
		}

	} else if len(os.Args) > 1 && os.Args[1] != "" {
		for c := range slice.Find(list, func(v Host) bool {
			return v.Name == os.Args[1] || v.Alias == os.Args[1]
		}).All() {
			log("cmd = %s", c.Cmd)
			if c.Cmd != "" {
				connect(c)
			} else {
				log("cmd is not defied")
			}
			os.Exit(0)

		}
		log("host not found %s", os.Args[1])
	}
}
