package skd

import (
	"errors"
	"math"
)

func CalculateExecutionTime(cpuMillis float64, curve Curve) float64 {
	return curve.a*math.Pow((cpuMillis/1000)+curve.b, -curve.c) + curve.d
}

func GetNumberOfPodSlots(nodeData *NodeData, numberOfPods int) (int, error) {
	initialNumberOfPodSlots := nodeData.initialNumberOfPodSlots
	if numberOfPods < initialNumberOfPodSlots {
		return initialNumberOfPodSlots, nil
	}
	updatedNumberOfPodSlots := initialNumberOfPodSlots
	for true {
		if updatedNumberOfPodSlots >= numberOfPods {
			return updatedNumberOfPodSlots, nil
		}
		updatedNumberOfPodSlots *= nodeData.podSlotScalingFactor
	}
	return -1, errors.New("should never happen")
}

func NodeContainsPod(nodeState NodeState, podName string) bool {
	for _, pod := range nodeState.pods {
		if pod.name == podName {
			return true
		}
	}
	return false
}

func GetCpuCoresPerPodAddingPods(nodeState NodeState, addingPods int) (float64, error) {
	if nodeState.node == nil {
		return -1, errors.New("nodeData is nil")
	}
	nodeData := nodeState.node
	numberOfPodSlots, err := GetNumberOfPodSlots(nodeData, len(nodeState.pods)+addingPods)
	if err != nil {
		return -1, err
	}
	return nodeData.allocatableCpu * nodeData.resourceLimit / float64(numberOfPodSlots), nil
}

func GetCpuCoresPerPod(nodeState NodeState) (float64, error) {
	return GetCpuCoresPerPodAddingPods(nodeState, 0)
}

func GetMemoryPerPodAddingPods(nodeState NodeState, addingPods int) (float64, error) {
	if nodeState.node == nil {
		return -1, errors.New("nodeData is nil")
	}
	nodeData := nodeState.node
	numberOfPodSlots, err := GetNumberOfPodSlots(nodeData, len(nodeState.pods)+addingPods)
	if err != nil {
		return -1, err
	}
	return nodeData.memory * nodeData.resourceLimit / float64(numberOfPodSlots), nil
}

func GetMemoryPerPod(nodeState NodeState) (float64, error) {
	return GetMemoryPerPodAddingPods(nodeState, 0)
}

func CalculatePenalty(state SystemState, networkPenalty float64, memoryPenalty float64) (float64, error) {
	var penalty = 0.0

	for _, nodeState := range state.nodes {
		cpuCoresPerPod, err := GetCpuCoresPerPod(nodeState)
		if err != nil {
			return -1, err
		}

		memoryPerPod, err := GetMemoryPerPod(nodeState)
		if err != nil {
			return -1, err
		}

		for _, podData := range nodeState.pods {
			penalty += math.Max(CalculateExecutionTime(cpuCoresPerPod, podData.curve)-podData.maximumExecutionTime, 0)
			for _, receivesDataFrom := range podData.receivesDataFrom {
				if !NodeContainsPod(nodeState, receivesDataFrom) {
					penalty += networkPenalty
				}
			}
			for _, dataSourceNodeName := range podData.dataSourceNodes {
				if nodeState.node.name != dataSourceNodeName {
					penalty += networkPenalty
				}
			}
			if memoryPerPod < podData.minimumMemory {
				penalty += memoryPenalty * (1 - memoryPerPod/podData.minimumMemory)
			}
		}
	}

	return penalty, nil
}