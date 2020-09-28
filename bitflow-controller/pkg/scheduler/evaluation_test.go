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

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldComparePenaltiesForSimpleCase() {
	var scheduler AdvancedScheduler
	curve := Curve{
		a: 6.71881241016441,
		b: 0.0486498280492762,
		c: 2.0417306475862214,
		d: 15.899403720950454,
	}

	scheduler = AdvancedScheduler{
		nodes: []*NodeData{
			{
				name:                    "n1",
				allocatableCpu:          4000,
				memory:                  1280,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n2",
				allocatableCpu:          4000,
				memory:                  640,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n3",
				allocatableCpu:          4000,
				memory:                  640,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
		},
		pods: []*PodData{
			{
				name:                 "p1",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p7", "p8"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p2",
				receivesDataFrom:     []string{"p10"},
				sendsDataTo:          []string{"p9"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p3",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p9"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p4",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p5",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p6",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p7",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p8",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p9",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p10",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p2"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p11",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p12",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p13",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
		},
		networkPenalty:                 200,
		memoryPenalty:                  1_000_000,
		executionTimePenaltyMultiplier: 2,
		thresholdPercent:               10,
	}

	start := time.Now()
	_, perfectSchedulingMap, errPerfectScheduling := scheduler.ScheduleCheckingAllPermutations()
	elapsedPerfectScheduling := time.Since(start)
	start = time.Now()
	_, actualSchedulingMap, errActualScheduling := scheduler.Schedule()
	elapsedActualScheduling := time.Since(start)

	s.Nil(errPerfectScheduling)
	s.Nil(errActualScheduling)

	perfectPenalty, errPerfectPenalty := scheduler.calculatePenaltyFromSchedulingMap(perfectSchedulingMap)
	actualPenalty, errActualPenalty := scheduler.calculatePenaltyFromSchedulingMap(actualSchedulingMap)

	s.Nil(errPerfectPenalty)
	s.Nil(errActualPenalty)

	println(fmt.Sprintf("perfect ms: %9d\tpenalty: %f", elapsedPerfectScheduling.Milliseconds(), perfectPenalty))
	println(fmt.Sprintf("actual ms:  %9d\tpenalty: %f", elapsedActualScheduling.Milliseconds(), actualPenalty))
}

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldComparePenaltiesForSimpleCase_variableMinimumMemory() {
	// Auswertung: Penalties var.memory

	var scheduler AdvancedScheduler
	curve := Curve{
		a: 6.71881241016441,
		b: 0.0486498280492762,
		c: 2.0417306475862214,
		d: 15.899403720950454,
	}

	scheduler = AdvancedScheduler{
		nodes: []*NodeData{
			{
				name:                    "n1",
				allocatableCpu:          4000,
				memory:                  1280,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n2",
				allocatableCpu:          4000,
				memory:                  640,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n3",
				allocatableCpu:          4000,
				memory:                  640,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
		},
		pods: []*PodData{
			{
				name:                 "p1",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p7", "p8"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p2",
				receivesDataFrom:     []string{"p10"},
				sendsDataTo:          []string{"p9"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p3",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p9"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p4",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p5",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p6",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p7",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p8",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p9",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p10",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p2"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p11",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p12",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p13",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
		},
		networkPenalty:                 200,
		memoryPenalty:                  1_000_000,
		executionTimePenaltyMultiplier: 2,
		thresholdPercent:               10,
	}

	println("minimumMemory;ms_perfect;penalty_perfect;ms_actual;penalty_actual")
	for minimumMemory := 1; minimumMemory <= 32; minimumMemory++ {
		for _, podData := range scheduler.pods {
			podData.minimumMemory = float64(minimumMemory)
		}

		start := time.Now()
		_, perfectSchedulingMap, errPerfectScheduling := scheduler.ScheduleCheckingAllPermutations()
		elapsedPerfectScheduling := time.Since(start)
		start = time.Now()
		_, actualSchedulingMap, errActualScheduling := scheduler.Schedule()
		elapsedActualScheduling := time.Since(start)

		s.Nil(errPerfectScheduling)
		s.Nil(errActualScheduling)

		perfectPenalty, errPerfectPenalty := scheduler.calculatePenaltyFromSchedulingMap(perfectSchedulingMap)
		actualPenalty, errActualPenalty := scheduler.calculatePenaltyFromSchedulingMap(actualSchedulingMap)

		s.Nil(errPerfectPenalty)
		s.Nil(errActualPenalty)

		println(fmt.Sprintf("%d;%d;%f;%d;%f", minimumMemory, elapsedPerfectScheduling.Milliseconds(), perfectPenalty, elapsedActualScheduling.Milliseconds(), actualPenalty))
	}
}

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldComparePenaltiesForSimpleCase_variableMaximumExecutionTime() {
	// Auswertung: Penalties var.executionTime

	var scheduler AdvancedScheduler
	curve := Curve{
		a: 6.71881241016441,
		b: 0.0486498280492762,
		c: 2.0417306475862214,
		d: 15.899403720950454,
	}

	scheduler = AdvancedScheduler{
		nodes: []*NodeData{
			{
				name:                    "n1",
				allocatableCpu:          4000,
				memory:                  1280,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n2",
				allocatableCpu:          4000,
				memory:                  640,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
			{
				name:                    "n3",
				allocatableCpu:          4000,
				memory:                  640,
				initialNumberOfPodSlots: 2,
				podSlotScalingFactor:    2,
				resourceLimit:           0.1,
			},
		},
		pods: []*PodData{
			{
				name:                 "p1",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p7", "p8"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p2",
				receivesDataFrom:     []string{"p10"},
				sendsDataTo:          []string{"p9"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p3",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p9"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p4",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p5",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p6",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p7",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p8",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p9",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p10",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p2"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p11",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p12",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p13",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
		},
		networkPenalty:                 1_000_000,
		memoryPenalty:                  1_000_000,
		executionTimePenaltyMultiplier: 2,
		thresholdPercent:               10,
	}

	println("maximumExecutionTime;ms_perfect;penalty_perfect;ms_actual;penalty_actual")
	for maximumExecutionTime := 50; maximumExecutionTime <= 3000; maximumExecutionTime += 50 {
		for _, podData := range scheduler.pods {
			podData.maximumExecutionTime = float64(maximumExecutionTime)
		}

		start := time.Now()
		_, perfectSchedulingMap, errPerfectScheduling := scheduler.ScheduleCheckingAllPermutations()
		elapsedPerfectScheduling := time.Since(start)
		start = time.Now()
		_, actualSchedulingMap, errActualScheduling := scheduler.Schedule()
		elapsedActualScheduling := time.Since(start)

		s.Nil(errPerfectScheduling)
		s.Nil(errActualScheduling)

		perfectPenalty, errPerfectPenalty := scheduler.calculatePenaltyFromSchedulingMap(perfectSchedulingMap)
		actualPenalty, errActualPenalty := scheduler.calculatePenaltyFromSchedulingMap(actualSchedulingMap)

		s.Nil(errPerfectPenalty)
		s.Nil(errActualPenalty)

		println(fmt.Sprintf("%d;%d;%f;%d;%f", maximumExecutionTime, elapsedPerfectScheduling.Milliseconds(), perfectPenalty, elapsedActualScheduling.Milliseconds(), actualPenalty))
	}
}

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldPrint3DGraphData() {
	maxNumberOfNodes := 10
	maxNumberOfPods := 10

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
			_, _, _ = scheduler.Schedule()
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
		print(fmt.Sprintf("%d;", measurement.executionTime.Milliseconds()))
		if measurement.numberOfPods == maxNumberOfPods {
			println()
		}
	}
}
