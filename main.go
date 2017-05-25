package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var metrics = []string{
	"loadavg5",
	"cpu.user.percentage",
	"cpu.nice.percentage",
	"cpu.system.percentage",
	"cpu.irq.percentage",
	"cpu.softirq.percentage",
	"cpu.iowait.percentage",
	"cpu.steal.percentage",
	"cpu.guest.percentage",
	"cpu.idle.percentage",
	"memory.used",
	"memory.buffers",
	"memory.cached",
	"memory.total",
	"memory.free",
	"interface.eth0.rxBytes.delta",
	"interface.eth0.txBytes.delta",
}

type (
	Host struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsRetired bool   `json:"isRetired"`
		LA
		CPU
		Memory
		Interface
	}
	LA struct {
		Avg5 float64
	}
	CPU struct {
		User    float64
		Nice    float64
		System  float64
		Irq     float64
		SoftIrq float64
		IoWait  float64
		Steal   float64
		Guest   float64
		Idle    float64
	}
	Memory struct {
		Used    float64
		Buffers float64
		Cached  float64
		Total   float64
		Free    float64
	}
	Interface struct {
		RxBytes float64
		TxBytes float64
	}
)

func getHosts(service, role string) ([]*Host, error) {
	var hs []*Host
	err := run("mkr", []string{"hosts", "--service", service, "--role", role}, &hs)
	return hs, err
}

func fetchMetrics(hosts []*Host, metricNames []string) (interface{}, error) {

	var v interface{}
	fetchArgs := []string{"fetch"}
	for _, m := range metricNames {
		fetchArgs = append(fetchArgs, []string{"-n", m}...)
	}
	for _, h := range hosts {
		if h.IsRetired == false {
			fetchArgs = append(fetchArgs, h.ID)
		}
	}

	if err := run("mkr", fetchArgs, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func run(command string, args []string, dst interface{}) error {
	var stdout bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	if dst != nil {
		r := bytes.NewReader(stdout.Bytes())
		if w, ok := dst.(io.Writer); ok {
			io.Copy(w, r)
		} else {
			if err := json.NewDecoder(r).Decode(dst); err != nil {
				return err
			}
		}
	}
	return nil
}

func tmpl() string {
	var t string
	now := "# {{ now }}"
	header := "Host        LoadAvg %CPU(user) %CPU(sys) %CPU(idle) Mem(total) Mem(used) Mem(buffers) Mem(cached) Mem(free) eth0(rxBytes) eth0(txBytes)"
	hostID := `{{ .ID }}`
	la := ` {{ .LA.Avg5 | cpuDigits }}`
	cpu := `     {{ .CPU.User | cpuDigits }}      {{ .CPU.System | cpuDigits }}      {{ .CPU.Idle | cpuDigits }}`
	mem := `     {{ .Memory.Total | memoryHumanReadable }}      {{ .Memory.Used | memoryHumanReadable }}     {{ .Memory.Buffers | memoryHumanReadable }}      {{ .Memory.Cached | memoryHumanReadable }}      {{ .Memory.Free | memoryHumanReadable }}`
	iface := `    {{ .Interface.RxBytes | trafficHumanReadable}}  {{ .Interface.TxBytes | trafficHumanReadable}}`

	t = now + "\n" + header + "\n" + `{{ range . }}` + hostID + la + cpu + mem + iface + "\n" + `{{ end }}`
	return t
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func trafficHumanReadable(size float64) string {
	var trafficUnits = []string{"B/s", "K/s", "M/s", "G/s", "T/s", "P/s", "E/s", "Z/s", "Y/s"}
	return humanReadable(size, trafficUnits)
}

func memoryHumanReadable(size float64) string {
	var memoryUnits = []string{"B", "K", "M", "G", "T", "P", "E", "Z", "Y"}
	return humanReadable(size, memoryUnits)
}

func cpuDigits(val float64) string {
	return strconv.FormatFloat(val, 'f', 2, 64)
}

func humanReadable(size float64, units []string) string {
	i := 0
	unitsLimit := len(units) - 1
	for size >= 1024.0 && i < unitsLimit {
		size = size / 1024.0
		i++
	}
	return fmt.Sprintf("%.2f%s", size, units[i])
}

func main() {

	var (
		service  string
		role     string
		interval int
	)
	flag.StringVar(&service, "s", "", "The service name(required). ")
	flag.StringVar(&role, "r", "", "The role name(required).")
	flag.IntVar(&interval, "i", 5, "The interval in second.")
	flag.Parse()

	if service == "" || role == "" {
		flag.PrintDefaults()
		return
	}

	for {
		hs, err := getHosts(service, role)
		if len(hs) == 0 {
			fmt.Printf("Couldn't find hosts in service(%s):role(%s)\n", service, role)
			return
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		v, err := fetchMetrics(hs, metrics)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, h := range hs {
			if w, ok := v.(map[string]interface{})[h.ID].(map[string]interface{}); ok {
				h.LA.Avg5 = w["loadavg5"].(map[string]interface{})["value"].(float64)

				h.CPU.User = w["cpu.user.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.Nice = w["cpu.nice.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.System = w["cpu.system.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.Irq = w["cpu.irq.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.SoftIrq = w["cpu.softirq.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.IoWait = w["cpu.iowait.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.Steal = w["cpu.steal.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.Guest = w["cpu.guest.percentage"].(map[string]interface{})["value"].(float64)
				h.CPU.Idle = w["cpu.idle.percentage"].(map[string]interface{})["value"].(float64)

				h.Memory.Used = w["memory.used"].(map[string]interface{})["value"].(float64)
				h.Memory.Buffers = w["memory.buffers"].(map[string]interface{})["value"].(float64)
				h.Memory.Cached = w["memory.cached"].(map[string]interface{})["value"].(float64)
				h.Memory.Total = w["memory.total"].(map[string]interface{})["value"].(float64)
				h.Memory.Free = w["memory.free"].(map[string]interface{})["value"].(float64)

				h.Interface.RxBytes = w["interface.eth0.rxBytes.delta"].(map[string]interface{})["value"].(float64)
				h.Interface.TxBytes = w["interface.eth0.txBytes.delta"].(map[string]interface{})["value"].(float64)
			}
		}

		t := template.New("t")
		t.Funcs(template.FuncMap{
			"now":                  nowString,
			"cpuDigits":            cpuDigits,
			"memoryHumanReadable":  memoryHumanReadable,
			"trafficHumanReadable": trafficHumanReadable,
		})
		template.Must(t.Parse(tmpl()))
		t.Execute(os.Stdout, hs)

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
