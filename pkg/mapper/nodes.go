package mapper

import (
	"fmt"
	"time"

	"github.com/leanix/leanix-k8s-connector/pkg/set"

	corev1 "k8s.io/api/core/v1"
)

func aggregrateNodes(nodes *corev1.NodeList) (map[string]interface{}, error) {
	nodeAggregate := make(map[string]interface{})
	items := nodes.Items
	if len(items) == 0 {
		return nodeAggregate, nil
	}
	availabilityZones := set.NewStringSet()
	nodeTypes := set.NewStringSet()
	architectures := set.NewStringSet()
	containerRuntimeVersion := set.NewStringSet()
	kernelVersion := set.NewStringSet()
	kubeletVersion := set.NewStringSet()
	operatingSystem := set.NewStringSet()
	osImage := set.NewStringSet()
	firstCreatedNode := items[0].ObjectMeta.CreationTimestamp
	lastCreatedNode := items[0].ObjectMeta.CreationTimestamp

	for _, n := range items {
		availabilityZones.Add(n.Labels["failure-domain.beta.kubernetes.io/zone"])
		nodeTypes.Add(n.Labels["beta.kubernetes.io/instance-type"])
		architectures.Add(n.Status.NodeInfo.Architecture)
		containerRuntimeVersion.Add(n.Status.NodeInfo.ContainerRuntimeVersion)
		kernelVersion.Add(n.Status.NodeInfo.KernelVersion)
		kubeletVersion.Add(n.Status.NodeInfo.KubeletVersion)
		operatingSystem.Add(n.Status.NodeInfo.OperatingSystem)
		osImage.Add(n.Status.NodeInfo.OSImage)
		nodeCreatedTimestamp := n.ObjectMeta.CreationTimestamp
		if nodeCreatedTimestamp.Before(&firstCreatedNode) {
			firstCreatedNode = nodeCreatedTimestamp
		}
		if nodeCreatedTimestamp.After(lastCreatedNode.Time) {
			lastCreatedNode = nodeCreatedTimestamp
		}
	}
	memory, err := aggregrateMemoryCapacity(&items)
	if err != nil {
		return nil, err
	}
	cpus, err := aggregrateCPUCapacity(&items)
	if err != nil {
		return nil, err
	}
	nodeAggregate["availabilityZones"] = availabilityZones.Items()
	nodeAggregate["dataCenter"] = items[0].Labels["failure-domain.beta.kubernetes.io/region"]
	nodeAggregate["nodeTypes"] = nodeTypes.Items()
	nodeAggregate["numberNodes"] = len(items)
	nodeAggregate["memoryCapacityGB"] = memory
	nodeAggregate["cpuCapacity"] = cpus
	nodeAggregate["architecture"] = architectures.Items()
	nodeAggregate["containerRuntimeVersion"] = containerRuntimeVersion.Items()
	nodeAggregate["kernelVersion"] = kernelVersion.Items()
	nodeAggregate["kubeletVersion"] = kubeletVersion.Items()
	nodeAggregate["operatingSystem"] = operatingSystem.Items()
	nodeAggregate["osImage"] = osImage.Items()
	nodeAggregate["labels"] = labelSet(&items)
	nodeAggregate["firstCreatedNode"] = firstCreatedNode.UTC().Format(time.RFC3339)
	nodeAggregate["lastCreatedNode"] = lastCreatedNode.UTC().Format(time.RFC3339)
	return nodeAggregate, nil
}

func aggregrateMemoryCapacity(nodes *[]corev1.Node) (float64, error) {
	var memoryCapacityGB float64
	for _, n := range *nodes {
		// The Memory() call returns the memory as resource.Quantity. 'Quantity is a fixed-point representation of a number.'
		// In order to calculate the memory capacity of all nodes, we get the bytes as int64 (hoping it does not exceed the int64 limit...).
		// We convert the bytes here to GiB to make sure that we do not exceed the limit of float64. This introduces a rounding error,
		// which we accept, because a precise value is not of interest for the user output.
		b, ok := n.Status.Capacity.Memory().AsInt64()
		if !ok {
			return 0, fmt.Errorf("Failed to get memory quantity as type int64")
		}
		memoryCapacityGB = memoryCapacityGB + byteToGiB(b)
	}
	return memoryCapacityGB, nil
}

func aggregrateCPUCapacity(nodes *[]corev1.Node) (int64, error) {
	var cpuCapacity int64
	for _, n := range *nodes {
		cores, ok := n.Status.Capacity.Cpu().AsInt64()
		if !ok {
			return 0, fmt.Errorf("Failed to get cpu quantity as type int64")
		}
		cpuCapacity = cpuCapacity + cores
	}
	return cpuCapacity, nil
}

func byteToGiB(b int64) float64 {
	return float64(b) / 1024 / 1024 / 1024
}

func labelSet(nodes *[]corev1.Node) map[string][]string {
	labelsAsSet := make(map[string]*set.String)
	labels := make(map[string][]string)
	for _, n := range *nodes {
		for l, v := range n.Labels {
			if labelsAsSet[l] == nil {
				labelsAsSet[l] = set.NewStringSet()
			}
			labelsAsSet[l].Add(v)
		}
	}
	for k, v := range labelsAsSet {
		labels[k] = v.Items()
	}
	return labels
}
