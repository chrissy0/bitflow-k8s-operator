package scheduler

import (
	"fmt"
	"github.com/bitflow-stream/bitflow-k8s-operator/bitflow-controller/pkg/common"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"time"
)

type EvaluationTestSuite struct {
	common.AbstractTestSuite
}

func TestEvaluation(t *testing.T) {
	suite.Run(t, new(EvaluationTestSuite))
}

func (s *EvaluationTestSuite) Test_AdvancedScheduler_shouldComparePenaltiesForSimpleCase() {
	curve := Curve{
		a: 6.71881241016441,
		b: 0.0486498280492762,
		c: 2.0417306475862214,
		d: 15.899403720950454,
	}

	var scheduler AdvancedScheduler
	scheduler = AdvancedScheduler{
		networkPenalty:                 200,
		memoryPenalty:                  1_000_000,
		executionTimePenaltyMultiplier: 2,
		thresholdPercent:               5,
		nodes: []*NodeData{
			{
				name:                    "n1",
				allocatableCpu:          4000,
				memory:                  320,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n2",
				allocatableCpu:          4000,
				memory:                  320,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n3",
				allocatableCpu:          4000,
				memory:                  320,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n4",
				allocatableCpu:          4000,
				memory:                  320,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
		},
		pods: []*PodData{
			{
				name:                 "p1",
				dataSourceNodes:      []string{"n1"},
				sendsDataTo:          []string{"p7", "p8"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p2",
				dataSourceNodes:      []string{"n2"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p3",
				dataSourceNodes:      []string{"n3"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p4",
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p5",
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p6",
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p7",
				receivesDataFrom:     []string{"p1"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p8",
				receivesDataFrom:     []string{"p1"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
		},
	}

	_, perfectSchedulingMap, errPerfectScheduling := scheduler.ScheduleCheckingAllPermutations()
	_, actualSchedulingMap, errActualScheduling := scheduler.Schedule()

	s.Nil(errPerfectScheduling)
	s.Nil(errActualScheduling)

	perfectPenalty, errPerfectPenalty := scheduler.calculatePenaltyFromSchedulingMap(perfectSchedulingMap)
	actualPenalty, errActualPenalty := scheduler.calculatePenaltyFromSchedulingMap(actualSchedulingMap)

	s.Nil(errPerfectPenalty)
	s.Nil(errActualPenalty)

	println(perfectPenalty)
	println(actualPenalty)
}

func (s *EvaluationTestSuite) Test_AdvancedScheduler_shouldPrint3DGraphData() {
	maxNumberOfNodes := 50
	maxNumberOfPods := 50

	type Measurement struct {
		numberOfNodes int
		numberOfPods  int
		executionTime time.Duration
	}

	var measurements []Measurement

	curve := Curve{
		a: 6.71881241016441,
		b: 0.0486498280492762,
		c: 2.0417306475862214,
		d: 15.899403720950454,
	}

	var scheduler AdvancedScheduler
	scheduler = AdvancedScheduler{
		networkPenalty:                 200,
		memoryPenalty:                  1_000_000,
		executionTimePenaltyMultiplier: 2,
		thresholdPercent:               5,
		nodes:                          []*NodeData{},
		pods:                           []*PodData{},
	}

	for numberOfNodes := 1; numberOfNodes <= maxNumberOfNodes; numberOfNodes++ {
		scheduler.pods = []*PodData{}
		scheduler.nodes = append(scheduler.nodes,
			&NodeData{
				name:                    "n" + strconv.Itoa(numberOfNodes),
				allocatableCpu:          4000,
				memory:                  320,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			})
		for numberOfPods := 1; numberOfPods <= maxNumberOfPods; numberOfPods++ {
			scheduler.pods = append(scheduler.pods,
				&PodData{
					name:                 "p" + strconv.Itoa(numberOfPods),
					dataSourceNodes:      []string{},
					sendsDataTo:          []string{},
					curve:                curve,
					minimumMemory:        16,
					maximumExecutionTime: 200,
				})

			println(fmt.Sprintf("measuring %d nodes x %d pods", numberOfNodes, numberOfPods))
			start := time.Now()
			scheduler.Schedule()
			elapsed := time.Since(start)

			measurements = append(measurements, Measurement{
				numberOfNodes: numberOfNodes,
				numberOfPods:  numberOfPods,
				executionTime: elapsed,
			})
		}
	}

	for i := 1; i <= maxNumberOfPods; i++ {
		print(";")
		print(i)
	}
	println()
	for _, measurement := range measurements {
		if measurement.numberOfPods == 1 {
			print(measurement.numberOfNodes)
			print(";")
		}
		print(fmt.Sprintf("%d;", measurement.executionTime.Nanoseconds()))
		if measurement.numberOfPods == maxNumberOfPods {
			println()
		}
	}
}
