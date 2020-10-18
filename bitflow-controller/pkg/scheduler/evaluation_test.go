package scheduler

import (
	"fmt"
	"github.com/bitflow-stream/bitflow-k8s-operator/bitflow-controller/pkg/common"
	"github.com/stretchr/testify/suite"
	"math"
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

	numberOfIterations := 15

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
				sendsDataTo:          []string{"p2", "p3", "p6", "p7", "p8", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p2",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{"p3", "p4", "p6", "p9", "p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p3",
				receivesDataFrom:     []string{"p1", "p2"},
				sendsDataTo:          []string{"p4", "p5", "p9", "p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p4",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{"p6", "p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p5",
				receivesDataFrom:     []string{"p3"},
				sendsDataTo:          []string{"p6", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p6",
				receivesDataFrom:     []string{"p1", "p2", "p4", "p5"},
				sendsDataTo:          []string{"p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p7",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p8",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p9",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{"p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p10",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p11",
				receivesDataFrom:     []string{"p2", "p3", "p4", "p6", "p9", "p10"},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p12",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p13",
				receivesDataFrom:     []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12"},
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

	println(fmt.Sprintf("%d iterations", numberOfIterations))
	println("minimumMemory;ns_perfect_avg;penalty_perfect_avg;ns_actual_avg;penalty_actual_avg")
	for minimumMemory := 1; minimumMemory <= 32; minimumMemory++ {
		for _, podData := range scheduler.pods {
			podData.minimumMemory = float64(minimumMemory)
		}

		var elapsedPerfectSchedulingTotalNs float64 = 0
		var elapsedActualSchedulingTotalNs float64 = 0.0
		var perfectPenaltyTotal float64 = 0.0
		var actualPenaltyTotal float64 = 0.0
		for i := 0; i < numberOfIterations; i++ {
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

			elapsedPerfectSchedulingTotalNs += float64(elapsedPerfectScheduling.Nanoseconds())
			elapsedActualSchedulingTotalNs += float64(elapsedActualScheduling.Nanoseconds())
			perfectPenaltyTotal += perfectPenalty
			actualPenaltyTotal += actualPenalty
		}
		println(fmt.Sprintf("%d;%f;%f;%f;%f",
			minimumMemory,
			elapsedPerfectSchedulingTotalNs/float64(numberOfIterations),
			perfectPenaltyTotal/float64(numberOfIterations),
			elapsedActualSchedulingTotalNs/float64(numberOfIterations),
			actualPenaltyTotal/float64(numberOfIterations)))
	}
}

func (s *EvaluationTestSuite) Test_AdvancedScheduler_shouldComparePenaltiesForSimpleCase_variableMaximumExecutionTime() {
	// Auswertung: Penalties var.maxExecutionTime

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
				sendsDataTo:          []string{"p2", "p3", "p6", "p7", "p8", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p2",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{"p3", "p4", "p6", "p9", "p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p3",
				receivesDataFrom:     []string{"p1", "p2"},
				sendsDataTo:          []string{"p4", "p5", "p9", "p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p4",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{"p6", "p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p5",
				receivesDataFrom:     []string{"p3"},
				sendsDataTo:          []string{"p6", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p6",
				receivesDataFrom:     []string{"p1", "p2", "p4", "p5"},
				sendsDataTo:          []string{"p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p7",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p8",
				receivesDataFrom:     []string{"p1"},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p9",
				receivesDataFrom:     []string{"p2", "p3"},
				sendsDataTo:          []string{"p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p10",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p11", "p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p11",
				receivesDataFrom:     []string{"p2", "p3", "p4", "p6", "p9", "p10"},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p12",
				receivesDataFrom:     []string{},
				sendsDataTo:          []string{"p13"},
				curve:                curve,
				minimumMemory:        16,
				maximumExecutionTime: 200,
			},
			{
				name:                 "p13",
				receivesDataFrom:     []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11", "p12"},
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

	println("maximumExecutionTime;ns_perfect;penalty_perfect;ns_actual;penalty_actual")
	for maximumExecutionTime := 50; maximumExecutionTime <= 3000; maximumExecutionTime += 10 {
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

		println(fmt.Sprintf("%d;%d;%f;%d;%f", maximumExecutionTime, elapsedPerfectScheduling.Nanoseconds(), perfectPenalty, elapsedActualScheduling.Nanoseconds(), actualPenalty))
	}
}

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldPrint3DGraphData() {
	maxNumberOfNodes := 10
	maxNumberOfPods := 10
	numberOfIterations := 10

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

	for iteration := 1; iteration <= numberOfIterations; iteration++ {
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
				println(fmt.Sprintf("i%d;n%d;p%d", iteration, numberOfNodes, numberOfPods))
				podData := &PodData{
					name:                 "p" + strconv.Itoa(numberOfPods),
					dataSourceNodes:      []string{},
					sendsDataTo:          []string{},
					receivesDataFrom:     []string{},
					curve:                curve,
					minimumMemory:        16,
					maximumExecutionTime: 200,
				}
				if numberOfPods == 1 {
					podData.dataSourceNodes = append(podData.dataSourceNodes, "n1")
				} else {
					podData.receivesDataFrom = append(podData.receivesDataFrom, "p"+strconv.Itoa(numberOfPods-1))
					scheduler.pods[len(scheduler.pods)-1].sendsDataTo = append(scheduler.pods[len(scheduler.pods)-1].sendsDataTo, "p"+strconv.Itoa(numberOfPods))
				}
				scheduler.pods = append(scheduler.pods, podData)

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
	}

	for i := 1; i <= maxNumberOfPods; i++ {
		print(";")
		print(i)
	}
	println()
	numberOfUniqueMeasurementSetups := len(measurements) / numberOfIterations
	for _, measurementIndex := range makeRange(0, numberOfUniqueMeasurementSetups-1) {
		measurement := measurements[measurementIndex]
		var accumulativeDuration time.Duration = 0
		for _, i := range makeRange(0, numberOfIterations-1) {
			currentMeasurement := measurements[measurementIndex+i*numberOfUniqueMeasurementSetups]
			accumulativeDuration += currentMeasurement.executionTime
		}
		measurement.executionTime = accumulativeDuration
		if measurement.numberOfPods == 1 {
			print(measurement.numberOfNodes)
			print(";")
		}
		print(fmt.Sprintf("%d;", measurement.executionTime.Nanoseconds()/int64(numberOfIterations)))
		if measurement.numberOfPods == maxNumberOfPods {
			println()
		}
	}
}

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldPrint3DGraphDataMoreConnections() {
	maxNumberOfNodes := 10
	maxNumberOfPods := 10
	numberOfIterations := 10

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

	for iteration := 1; iteration <= numberOfIterations; iteration++ {
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
				println(fmt.Sprintf("i%d;n%d;p%d", iteration, numberOfNodes, numberOfPods))
				podData := &PodData{
					name:                 "p" + strconv.Itoa(numberOfPods),
					dataSourceNodes:      []string{},
					sendsDataTo:          []string{},
					receivesDataFrom:     []string{},
					curve:                curve,
					minimumMemory:        16,
					maximumExecutionTime: 200,
				}
				if numberOfPods == 1 {
					podData.dataSourceNodes = append(podData.dataSourceNodes, "n1")
				} else {
					podData.receivesDataFrom = append(podData.receivesDataFrom, "p"+strconv.Itoa(numberOfPods-1))
					scheduler.pods[len(scheduler.pods)-1].sendsDataTo = append(scheduler.pods[len(scheduler.pods)-1].sendsDataTo, "p"+strconv.Itoa(numberOfPods))
					if numberOfPods >= 4 {
						connectedTo := math.Round(float64(numberOfPods) / 2.0)
						podData.receivesDataFrom = append(podData.receivesDataFrom, "p"+strconv.Itoa(int(connectedTo)))
						scheduler.pods[int64(connectedTo)-1].sendsDataTo = append(scheduler.pods[int64(connectedTo)-1].sendsDataTo, "p"+strconv.Itoa(numberOfPods))
					}
				}
				scheduler.pods = append(scheduler.pods, podData)

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
	}

	for i := 1; i <= maxNumberOfPods; i++ {
		print(";")
		print(i)
	}
	println()
	numberOfUniqueMeasurementSetups := len(measurements) / numberOfIterations
	for _, measurementIndex := range makeRange(0, numberOfUniqueMeasurementSetups-1) {
		measurement := measurements[measurementIndex]
		var accumulativeDuration time.Duration = 0
		for _, i := range makeRange(0, numberOfIterations-1) {
			currentMeasurement := measurements[measurementIndex+i*numberOfUniqueMeasurementSetups]
			accumulativeDuration += currentMeasurement.executionTime
		}
		measurement.executionTime = accumulativeDuration
		if measurement.numberOfPods == 1 {
			print(measurement.numberOfNodes)
			print(";")
		}
		print(fmt.Sprintf("%d;", measurement.executionTime.Nanoseconds()/int64(numberOfIterations)))
		if measurement.numberOfPods == maxNumberOfPods {
			println()
		}
	}
}

func (s *EvaluationTestSuite) xTest_AdvancedScheduler_shouldIterateOverNumber() {
	maxNumberOfNodes := 200
	totalNumberOfPods := 200

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

	for numberOfPods := 1; numberOfPods <= totalNumberOfPods; numberOfPods++ {
		podData := &PodData{
			name:                 "p" + strconv.Itoa(numberOfPods),
			dataSourceNodes:      []string{},
			sendsDataTo:          []string{},
			receivesDataFrom:     []string{},
			curve:                curve,
			minimumMemory:        16,
			maximumExecutionTime: 200,
		}
		scheduler.pods = append(scheduler.pods, podData)
	}

	for numberOfNodes := 1; numberOfNodes <= maxNumberOfNodes; numberOfNodes++ {
		scheduler.nodes = []*NodeData{}
		for currentNumberOfNodes := 1; currentNumberOfNodes <= numberOfNodes; currentNumberOfNodes++ {
			scheduler.nodes = append(scheduler.nodes,
				&NodeData{
					name:                    "n" + strconv.Itoa(numberOfNodes),
					allocatableCpu:          4000,
					memory:                  320,
					initialNumberOfPodSlots: 2,
					podSlotScalingFactor:    2,
					resourceLimit:           0.1,
				})
		}

		print(numberOfNodes)
		print(";")
		_, _, _ = scheduler.Schedule()

	}
}
