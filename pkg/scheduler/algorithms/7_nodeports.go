package algorithms

import (
	"fmt"
	"net"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (a *AlgoManager) NodePorts(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodeNameList := make([]string, 0)
	fmt.Printf("pod add -> name : %s\n", args.Pod.Name)
	fmt.Println("** NodePorts Algorithm **")
	if args.Pod.Spec.HostNetwork {
		nodeIPMap := make(map[string]string)
		nodePortMap := make(map[string]bool)
		portList := make([]string, 0)
		for _, container := range args.Pod.Spec.Containers {
			for _, port := range container.Ports {
				portList = append(portList, strconv.Itoa(int(port.ContainerPort)))
			}
		}
		fmt.Print("Pod request port : ")

		for i, port := range portList {
			if i == len(portList)-1 {
				fmt.Printf("%s\n", port)
			} else {
				fmt.Printf("%s ,", port)
			}
		}
		for _, node := range args.Nodes.Items {
			nodePortMap[node.Name] = true
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					nodeIPMap[node.Name] = addr.Address
				}
			}
			for _, port := range portList {
				if a.portCheck(net.JoinHostPort(nodeIPMap[node.Name], port)) {
					nodePortMap[node.Name] = false
					break
				} else {
					continue
				}
			}
		}
		for nodeName, isPortAvaliable := range nodePortMap {
			if isPortAvaliable {
				nodeNameList = append(nodeNameList, nodeName)
				fmt.Println(nodeName, " : Available")
			} else {
				fmt.Println(nodeName, " : Unavailable")
				continue
			}
		}

		fmt.Println("Selected Node(s) : ", nodeNameList)
		return &extenderv1.ExtenderFilterResult{NodeNames: &nodeNameList}, nil
	} else {
		fmt.Printf("Selected Node(s) : %s\n", *args.NodeNames)
		return &extenderv1.ExtenderFilterResult{NodeNames: args.NodeNames}, nil
	}
}

func (a *AlgoManager) portCheck(hostport string) bool {
	conn, err := net.DialTimeout("tcp", hostport, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
