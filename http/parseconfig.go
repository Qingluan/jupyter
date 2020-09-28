package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	// "github.com/rs/zerolog/log"
)

type EnumeConfig struct {
	Domain   string
	Proxy    string
	Output   string
	Names    []string
	StartId  int
	EndId    int
	Template map[string]string
}

var (
	ConfigDocument = `
## this is demo config file
domain: https://www.dz3.5.com/?id={id}
proxy: socks5://127.0.0.1:1091
startId: 0
endId : 100
output: test.csv


------------------- TEMP EXTRACT -------------------
{
	"names": ["uid","name","regist time","group", "reply"],
	"css":{
		"uid":         "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(1) > h2 > span",
		"name":        "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(1) > h2 ",
		"reply":       "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(1) > ul.cl.bbda.pbm.mbm > li > a:nth-child(4)",
		"group":       "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(2) > ul:nth-child(2) > li:nth-child(2) > span > a",
		"regist time": "#pbbs > li:nth-child(2)"
	}
}
`
)

func ReadConf(f string) (config *EnumeConfig) {
	fp, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	config = new(EnumeConfig)
	readyJson := false
	templateS := ""
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") && !readyJson {
			fs := strings.SplitN(line, ":", 2)
			head, body := strings.TrimSpace(fs[0]), strings.TrimSpace(fs[1])
			switch head {
			case "domain":
				log.Println(fmt.Sprintf("[%s]: %s", head, body))
				config.Domain = body
			case "startId":
				log.Println(fmt.Sprintf("[%s]: %s", head, body))

				config.StartId, _ = strconv.Atoi(body)
			case "endId":
				config.EndId, _ = strconv.Atoi(body)
				log.Println(fmt.Sprintf("[%s]: %s", head, body))
			case "proxy":
				log.Println(fmt.Sprintf("[%s]: %s", head, body))
				config.Proxy = body
			case "output":
				log.Println(fmt.Sprintf("[%s]: %s", head, body))
				config.Output = body
			}
		} else if strings.HasPrefix(line, "------------------- TEMP EXTRACT ") {
			readyJson = true
		} else if readyJson {
			templateS += line
		}
	}
	ms := make(map[string]interface{})
	config.Template = make(map[string]string)
	json.Unmarshal([]byte(templateS), &ms)
	for _, i := range ms["names"].([]interface{}) {
		config.Names = append(config.Names, i.(string))
	}
	for k, v := range ms["css"].(map[string]interface{}) {
		config.Template[k] = v.(string)
	}
	return
}

func ShowDemo() {
	fmt.Println(ConfigDocument)
}

func (conf *EnumeConfig) Marshal() string {
	b, _ := json.MarshalIndent(conf, "", "\t")
	return string(b)
}
