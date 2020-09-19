package skd

import (
	"errors"
	"fmt"
)

// TODO Error message am Ende: X Pods haben zu wenig CPU, Y Pods haben zu wenig Memory

type Scheduler interface {
	Schedule() (bool, map[string]string, error)
}

type EqualDistributionScheduler struct {
	nodeNames []string
	podNames  []string
}

var calculationCount = 0

func (eds EqualDistributionScheduler) Schedule() (bool, map[string]string, error) {
	if err := validateEqualDistributionScheduler(eds); err != nil {
		return false, nil, err
	}

	m := make(map[string]string)

	if len(eds.podNames) == 0 {
		return true, m, nil
	}

	var nodeIndex = 0
	for _, podName := range eds.podNames {
		if nodeIndex >= len(eds.nodeNames) {
			nodeIndex = 0
		}

		m[podName] = eds.nodeNames[nodeIndex]

		nodeIndex++
	}

	return true, m, nil
}

// TODO make available outside of package
type AdvancedScheduler struct {
	nodes              []*NodeData
	pods               []*PodData
	networkPenalty     float64
	memoryPenalty      float64
	thresholdPercent   float64
	previousScheduling map[string]string
}

func sortPodsUsingKahnsAlgorithm(pods []*PodData) ([]PodData, error) {
	var initialPods []PodData
	for _, pod := range pods {
		podCopy := PodData{
			// TODO remove TODO at very end -> did fields get added to PodData? Add here too.
			name:                 pod.name,
			curve:                pod.curve,
			dataSourceNodes:      make([]string, len(pod.dataSourceNodes)),
			receivesDataFrom:     make([]string, len(pod.receivesDataFrom)),
			sendsDataTo:          make([]string, len(pod.sendsDataTo)),
			minimumMemory:        pod.minimumMemory,
			maximumExecutionTime: pod.maximumExecutionTime,
		}
		copy(podCopy.dataSourceNodes, pod.dataSourceNodes)
		copy(podCopy.receivesDataFrom, pod.receivesDataFrom)
		copy(podCopy.sendsDataTo, pod.sendsDataTo)
		initialPods = append(initialPods, podCopy)
	}

	sortedPodNames := []string{}
	noIncomingEdgePods := []*PodData{}

	for _, pod := range pods {
		if len(pod.receivesDataFrom) == 0 {
			noIncomingEdgePods = append(noIncomingEdgePods, pod)
		}
	}

	for len(noIncomingEdgePods) != 0 {
		newlySortedPod := noIncomingEdgePods[0]
		noIncomingEdgePods = noIncomingEdgePods[1:]

		sortedPodNames = append(sortedPodNames, newlySortedPod.name)

		for _, receiverPodName := range newlySortedPod.sendsDataTo {
			var receiverPod *PodData
			for _, potentialReceiverPod := range pods {
				if potentialReceiverPod.name == receiverPodName {
					receiverPod = potentialReceiverPod
					break
				}
			}
			if receiverPod == nil || receiverPod.name == "" || receiverPod.name != receiverPodName || receiverPod.receivesDataFrom == nil || len(receiverPod.receivesDataFrom) == 0 {
				return nil, errors.New(fmt.Sprintf("pod %s is referenced but does not exist or is missing correct receivesDataFrom entry", receiverPodName))
			}
			// remove edge (both ways)
			newlySortedPod.sendsDataTo = newlySortedPod.sendsDataTo[1:]
			for i, senderPodName := range receiverPod.receivesDataFrom {
				if senderPodName == newlySortedPod.name {
					// Remove the element at index i from receivedDataFrom
					copy(receiverPod.receivesDataFrom[i:], receiverPod.receivesDataFrom[i+1:])                        // Shift a[i+1:] left one index.
					receiverPod.receivesDataFrom[len(receiverPod.receivesDataFrom)-1] = ""                            // Erase last element (write zero value).
					receiverPod.receivesDataFrom = receiverPod.receivesDataFrom[:len(receiverPod.receivesDataFrom)-1] // Truncate slice.
					break
				}
			}
			if len(receiverPod.receivesDataFrom) == 0 {
				noIncomingEdgePods = append(noIncomingEdgePods, receiverPod)
			}
		}
	}

	for _, pod := range pods {
		if len(pod.receivesDataFrom) != 0 || len(pod.sendsDataTo) != 0 {
			return nil, errors.New(fmt.Sprintf("Pod %s has edge after sorting, make sure there is a 'receivesDataFrom' entry for every 'sendsDataTo' entry. Does the graph have a cycle?", pod.name))
		}
	}
	sortedPods := []PodData{}
	for _, podName := range sortedPodNames {
		for _, pod := range initialPods {
			if pod.name == podName {
				sortedPods = append(sortedPods, pod)
				break
			}
		}
	}
	return sortedPods, nil
}

func (as AdvancedScheduler) findBestSchedulingCheckingAllPermutations(state SystemState, podsLeft []*PodData) (SystemState, float64, error) {
	if len(podsLeft) == 0 {
		penalty, err := CalculatePenalty(state, as.networkPenalty, as.memoryPenalty)
		calculationCount++
		return state, penalty, err
	}

	currentPod := podsLeft[0]

	var lowestPenalty float64 = -1
	var lowestPenaltySystemState SystemState

	for i, nodeState := range state.nodes {
		nodeState.pods = append(nodeState.pods, currentPod)
		state.nodes[i] = nodeState
		newSystemState, currentPenalty, err := as.findBestSchedulingCheckingAllPermutations(state, podsLeft[1:])
		if err != nil {
			state.nodes[i].pods = removeLastPodFromSlice(state.nodes[i].pods)
			continue
		}
		if lowestPenalty == -1 || currentPenalty < lowestPenalty {
			lowestPenalty = currentPenalty

			// copying "manually" to prevent lowestPenaltySystemState and newSystemState from having the same memory address, which leads to problems
			lowestPenaltySystemState = SystemState{}
			for _, newSystemStateNodeState := range newSystemState.nodes {
				lowestPenaltySystemState.nodes = append(lowestPenaltySystemState.nodes, newSystemStateNodeState)
			}
		}

		state.nodes[i].pods = removeLastPodFromSlice(state.nodes[i].pods)
	}

	if lowestPenalty == -1 {
		return SystemState{}, -1, errors.New("pod " + currentPod.name + " could not be scheduled onto any node")
	}

	return lowestPenaltySystemState, lowestPenalty, nil
}

func (as AdvancedScheduler) findGoodScheduling(state SystemState, podsLeft []*PodData, currentlyLowestPenalty float64) (SystemState, float64, error) {
	if len(podsLeft) == 0 {
		penalty, err := CalculatePenalty(state, as.networkPenalty, as.memoryPenalty)
		calculationCount++
		return state, penalty, err
	} else if currentlyLowestPenalty != -1 {
		penalty, err := CalculatePenalty(state, as.networkPenalty, as.memoryPenalty)
		// TODO score might be lowered by added pods which communicate (removing network-penalty), needs to be taken into account
		if err != nil || penalty > currentlyLowestPenalty {
			return state, penalty, errors.New("permutation does not have lower penalty, skipping")
		}
	}

	currentPod := podsLeft[0]

	var lowestPenalty float64 = -1
	var lowestPenaltySystemState SystemState

	for i, nodeState := range state.nodes {
		nodeState.pods = append(nodeState.pods, currentPod)
		state.nodes[i] = nodeState
		newSystemState, currentPenalty, err := as.findGoodScheduling(state, podsLeft[1:], lowestPenalty)
		if err != nil {
			state.nodes[i].pods = removeLastPodFromSlice(state.nodes[i].pods)
			continue
		}
		if lowestPenalty == -1 || currentPenalty < lowestPenalty {
			lowestPenalty = currentPenalty

			// copying "manually" to prevent lowestPenaltySystemState and newSystemState from having the same memory address, which leads to problems
			lowestPenaltySystemState = SystemState{}
			for _, newSystemStateNodeState := range newSystemState.nodes {
				lowestPenaltySystemState.nodes = append(lowestPenaltySystemState.nodes, newSystemStateNodeState)
			}
		}

		state.nodes[i].pods = removeLastPodFromSlice(state.nodes[i].pods)
	}

	if lowestPenalty == -1 {
		return SystemState{}, -1, errors.New("pod " + currentPod.name + " could not be scheduled onto any node")
	}

	return lowestPenaltySystemState, lowestPenalty, nil
}

func removeLastPodFromSlice(pods []*PodData) []*PodData {
	// deleting previously added pod in preparation for next iteration
	// copying "manually" to prevent errors -  pods = pods[:len(pods)-1] does NOT work
	var tempPods []*PodData
	for j, pod := range pods {
		if j == len(pods)-1 {
			break
		}
		tempPods = append(tempPods, pod)
	}
	return tempPods
}

func NewDistributionPenaltyLowerConsideringThreshold(previousPenalty float64, newPenalty float64, thresholdPercent float64) bool {
	var previousPenaltyMinusThreshold = previousPenalty * ((100 - thresholdPercent) / 100)
	if newPenalty <= previousPenaltyMinusThreshold {
		return true
	}
	return false
}

func getSystemStateFromSchedulingMap(nodes []*NodeData, pods []*PodData, scheduling map[string]string) SystemState {
	systemState := SystemState{[]NodeState{}}

	for _, node := range nodes {
		nodeState := NodeState{
			node: node,
			pods: []*PodData{},
		}
		for _, pod := range pods {
			if scheduling[pod.name] == node.name {
				nodeState.pods = append(nodeState.pods, pod)
			}
		}
		systemState.nodes = append(systemState.nodes, nodeState)
	}

	return systemState
}

func (as AdvancedScheduler) getPreviousSystemState() SystemState {
	return getSystemStateFromSchedulingMap(as.nodes, as.pods, as.previousScheduling)
}

func (as AdvancedScheduler) Schedule() (bool, map[string]string, error) {
	if err := validateAdvancedScheduler(as); err != nil {
		return false, nil, err
	}

	systemState := SystemState{[]NodeState{}}
	for _, node := range as.nodes {
		systemState.nodes = append(systemState.nodes, NodeState{
			node: node,
			pods: []*PodData{},
		})
	}

	calculationCount = 0

	//bestDistributionState, bestDistributionPenalty, err := as.findBestSchedulingCheckingAllPermutations(systemState, as.pods)
	bestDistributionState, bestDistributionPenalty, err := as.findGoodScheduling(systemState, as.pods, -1)

	if as.previousScheduling != nil {
		previousPenalty, err := CalculatePenalty(as.getPreviousSystemState(), as.networkPenalty, as.memoryPenalty)
		if err == nil && !NewDistributionPenaltyLowerConsideringThreshold(previousPenalty, bestDistributionPenalty, as.thresholdPercent) {
			return false, nil, nil
		}
	}

	if err != nil {
		return false, nil, err
	}

	m := make(map[string]string)
	for _, nodeState := range bestDistributionState.nodes {
		nodeName := nodeState.node.name
		for _, pod := range nodeState.pods {
			m[pod.name] = nodeName
		}
	}

	println(fmt.Sprintf("Penalty: %f\nCalculations: %d", bestDistributionPenalty, calculationCount))
	return true, m, nil
}

type NodeData struct {
	name                    string
	allocatableCpu          float64 // 1000 == 1 CPU core
	memory                  float64 // memory in MB
	initialNumberOfPodSlots int
	podSlotScalingFactor    int
	resourceLimit           float64
}

type PodData struct {
	name                 string
	dataSourceNodes      []string // list of node names
	receivesDataFrom     []string // list of pod names
	sendsDataTo          []string // list of pod names TODO necessary?
	curve                Curve
	minimumMemory        float64 // memory in MB
	maximumExecutionTime float64 // maximum execution time in ms
}

type Curve struct {
	a, b, c, d float64
}

type NodeState struct {
	node *NodeData
	pods []*PodData
}

func (state SystemState) toString() string {
	var str = "("

	for _, nodeState := range state.nodes {
		str += nodeState.node.name + "["

		for _, pod := range nodeState.pods {
			str += pod.name + " "
		}

		str += "] "
	}
	str += ")"
	return str
}

type SystemState struct {
	nodes []NodeState
}

func validateEqualDistributionScheduler(scheduler EqualDistributionScheduler) error {
	if len(scheduler.nodeNames) == 0 {
		return errors.New("no nodes in scheduler")
	}
	for _, name := range scheduler.nodeNames {
		if name == "" {
			return errors.New("empty name in nodeNames")
		}
	}
	if len(scheduler.podNames) == 0 {
		return errors.New("no pods in scheduler")
	}
	for _, name := range scheduler.podNames {
		if name == "" {
			return errors.New("empty name in podNames")
		}
	}
	return nil
}

func validateAdvancedScheduler(scheduler AdvancedScheduler) error {
	if len(scheduler.nodes) == 0 {
		return errors.New("no node data in scheduler")
	}
	for _, nodeData := range scheduler.nodes {
		if nodeData.name == "" {
			return errors.New("empty name in NodeData")
		}
		if nodeData.memory <= 0 {
			return errors.New("memory is <= 0")
		}
		if nodeData.initialNumberOfPodSlots <= 0 {
			return errors.New("initialNumberOfPodSlots is <= 0")
		}
		if nodeData.podSlotScalingFactor <= 0 {
			return errors.New("podSlotScalingFactor is <= 0")
		}
		if nodeData.resourceLimit <= 0 {
			return errors.New("resourceLimit is <= 0")
		}
	}
	for _, podData := range scheduler.pods {
		if podData.name == "" {
			return errors.New("empty name in PodData")
		}
		if podData.receivesDataFrom == nil {
			return errors.New("receivesDataFrom is nil")
		}
		if podData.curve == (Curve{}) {
			return errors.New("empty curve")
		}
		if podData.minimumMemory <= 0 {
			return errors.New("minimumMemory is <= 0")
		}
	}
	if scheduler.thresholdPercent < 0 || scheduler.thresholdPercent > 100 {
		return errors.New("thresholdPercent needs to be >= 0 and <= 100")
	}
	return nil
}
