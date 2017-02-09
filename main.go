package main

import (
	"flag"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"log"
	"os"
	"strconv"
)

func main() {
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	flag.Parse()

	c, _ := proxmox.NewClient(os.Getenv("PM_API_URL"), nil, nil)
	err := c.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"))
	if err != nil {
		log.Fatal(err)
	}

	vmid, _ := strconv.Atoi(flag.Args()[1])

	var jbody interface{}
	var vmr *proxmox.VmRef
	switch flag.Args()[0] {
	case "start":
		vmr = proxmox.NewVmRef(vmid)
		jbody, _ = c.StartVm(vmr)
	case "stop":
		vmr = proxmox.NewVmRef(vmid)
		jbody, _ = c.StopVm(vmr)
	case "createQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		failError(err)
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		failError(config.CreateVm(vmr, c))
	case "installQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		maxid, err := proxmox.MaxVmId(c)
		failError(err)
		vmr = proxmox.NewVmRef(maxid + 1)
		vmr.SetNode(flag.Args()[1])
		log.Print("Creating node: ")
		log.Println(vmr)
		failError(config.CreateVm(vmr, c))
		_, err = c.StartVm(vmr)
		failError(err)
		sshPort, err := proxmox.SshForwardUsernet(vmr, c)
		failError(err)
		log.Println("Waiting for CDRom install shutdown (at least 5 minutes)")
		failError(proxmox.WaitForShutdown(vmr, c))
		log.Println("Restarting")
		_, err = c.StartVm(vmr)
		failError(err)
		sshPort, err = proxmox.SshForwardUsernet(vmr, c)
		failError(err)
		log.Println("SSH Portforward on:" + sshPort)
		log.Println("Complete")
	case "sshforward":
		vmr = proxmox.NewVmRef(vmid)
		sshPort, err := proxmox.SshForwardUsernet(vmr, c)
		failError(err)
		log.Println("SSH Portforward on:" + sshPort)
	default:
		fmt.Printf("unknown action, try start|stop vmid")
	}

	log.Println(jbody)
	//log.Println(vmr)
}

func failError(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}