import {
  currentDataSourcesMap,
  currentGraphElementsWithStacksMap,
  currentPodsMap,
  currentStepsMap,
  DataSource,
  dataSourceMap,
  DataSourceStack,
  GraphElement,
  Pod,
  podMap,
  PodStack,
  Step,
  stepMap
} from "../definitions/definitions";
import {uuidv4} from "../util/util";
import {maxNumberOfSeparateGraphElements} from "../config/config";

export function getAllDataSources(): DataSource[] {
  return Array.from(dataSourceMap.keys()).map(dataSourceKey => <DataSource>dataSourceMap.get(dataSourceKey)).filter(dataSource => dataSource != undefined);
}

export function getAllSteps(): Step[] {
  return Array.from(stepMap.keys()).map(stepKey => <Step>stepMap.get(stepKey)).filter(step => step != undefined);
}

export function getAllPods(): Pod[] {
  return Array.from(podMap.keys()).map(podKey => <Pod>podMap.get(podKey)).filter(pod => pod != undefined);
}

export function getCurrentDataSources(): DataSource[] {
  return Array.from(currentDataSourcesMap.keys()).map(dataSourceKey => <DataSource>currentDataSourcesMap.get(dataSourceKey)).filter(dataSource => dataSource != undefined);
}

export function getCurrentSteps(): Step[] {
  return Array.from(currentStepsMap.keys()).map(stepKey => <Step>currentStepsMap.get(stepKey)).filter(step => step != undefined);
}

export function getCurrentPods(): Pod[] {
  return Array.from(currentPodsMap.keys()).map(podKey => <Pod>currentPodsMap.get(podKey)).filter(pod => pod != undefined);
}

export function setCurrentDataSources(dataSources: DataSource[]) {
  currentDataSourcesMap.clear();
  dataSources.forEach(dataSource => {
    currentDataSourcesMap.set(dataSource.name, dataSource);
  });
}

export function setCurrentSteps(steps: Step[]) {
  currentStepsMap.clear();
  steps.forEach(step => {
    currentStepsMap.set(step.name, step);
  });
}

export function setCurrentPods(pods: Pod[]) {
  currentPodsMap.clear();
  pods.forEach(pod => {
    currentPodsMap.set(pod.name, pod);
  });
}

export function getDepthOfDataSource(elementName: string): number {
  let element = dataSourceMap.get(elementName);
  if (element == undefined) {
    return 0;
  }
  if (!element.hasCreatorPod) {
    return 0;
  }
  let depth = getDepthOfPod(element.creatorPod.name);
  if (depth == undefined) {
    depth = 0;
  }
  return depth + 1;
}

export function getDepthOfPod(podName: string): number {
  let element = podMap.get(podName);
  if (element == undefined) {
    return 0;
  }
  if (element.creatorDataSources == undefined || element.creatorDataSources.length === 0 || !element.hasCreatorStep) {
    return 0;
  }

  return 1 + element.creatorDataSources.map(dataSource => {
    let depth: number = getDepthOfDataSource(dataSource.name);
    if (depth == undefined) {
      return 0;
    }
    return depth;
  }).reduce((p, c) => {
    if (c == undefined) {
      return p;
    }
    return Math.max(p, c)
  });
}

export function getDepthOfStep(stepName: string): number {
  let element = stepMap.get(stepName);
  if (element == undefined) {
    return 0;
  }
  if ((element.podType === 'pod' && (element.pods == undefined || element.pods.length === 0)) ||
    (element.podType === 'pod-stack' && (element.podStack == undefined || element.podStack.pods == undefined || element.podStack.pods.length === 0))) {
    return 0;
  }
  let depth: number = 0;
  let podNames: string[];
  if (element.podType === 'pod') {
    podNames = element.pods.map(pod => pod.name);
  } else if (element.podType === 'pod-stack') {
    podNames = element.podStack.pods.map(pod => pod.name);
  } else {
    return 0;
  }
  podNames.forEach(podName => {
    let podDepth = getDepthOfPod(podName);
    if (podDepth != undefined && podDepth > depth) {
      depth = podDepth;
    }
  });
  return depth;
}

export function getDepthOfDataSourceStack(dataSourceStack: DataSourceStack): number {
  return dataSourceStack.dataSources.map(dataSource => {
    let depth: number | undefined = getDepthOfDataSource(dataSource.name);
    if (depth == undefined) {
      return 0;
    }
    return depth;
  }).reduce((p, c) => {
    if (c == undefined) {
      return p;
    }
    return Math.max(p, c)
  });
}

export function getDepthOfPodStack(podStack: PodStack): number {
  return podStack.pods.map(pod => {
    let depth: number | undefined = getDepthOfPod(pod.name);
    if (depth == undefined) {
      return 0;
    }
    return depth;
  }).reduce((p, c) => {
    if (c == undefined) {
      return p;
    }
    return Math.max(p, c)
  });
}

export function getDepthOfGraphElement(graphElement: GraphElement) {
  switch (graphElement.type) {
    case "step":
      return getDepthOfStep(graphElement.step.name);
    case "data-source":
      return getDepthOfDataSource(graphElement.dataSource.name);
    case "pod":
      return getDepthOfPod(graphElement.pod.name);
    case "data-source-stack":
      return getDepthOfDataSourceStack(graphElement.dataSourceStack);
    case "pod-stack":
      return getDepthOfPodStack(graphElement.podStack);
    default:
      return 0;
  }
}


export function addCreatorPodsToDataSources() {
  getAllDataSources().forEach(dataSource => {
    if (dataSource.hasCreatorPod) {
      dataSource.creatorPod = podMap.get(dataSource.creatorPod.name);
    }
  });
}

export function setCurrentGraphElements(dataSources, steps, pods) {
  setCurrentDataSources(dataSources);
  setCurrentSteps(steps);
  setCurrentPods(pods);
}

export function podShouldBeGroupedWithPodStack(pod: Pod, podStack: PodStack) {
  return (!pod.hasCreatorStep && !podStack.hasCreatorStep) || (pod.hasCreatorStep && podStack.hasCreatorStep && pod.creatorStep.name === podStack.creatorStep.name);
}

function arraysEqual(_arr1, _arr2) {
  if (!Array.isArray(_arr1) || !Array.isArray(_arr2) || _arr1.length !== _arr2.length)
    return false;

  var arr1 = _arr1.concat().sort();
  var arr2 = _arr2.concat().sort();

  for (var i = 0; i < arr1.length; i++) {
    if (arr1[i] !== arr2[i])
      return false;
  }
  return true;
}

function getCreatedPodsFromDataSourceStack(dataSourceStack: DataSourceStack): Pod[] {
  let pods = [];
  dataSourceStack.dataSources.forEach(dataSource => {
    pods = [...pods, ...dataSource.createdPods];
  });
  return [...new Set(pods)];
}

function podsToPodAndPodCreatorStepNames(pods: Pod[]) {
  let podAndPodStackNames = [];
  pods.forEach(pod => {
    if (pod.creatorStep == undefined) {
      podAndPodStackNames = [...podAndPodStackNames, ...[pod.name]];
    } else {
      podAndPodStackNames = [...podAndPodStackNames, ...[pod.creatorStep.name]];
    }
  });
  return [...new Set(podAndPodStackNames)];
}

export function dataSourceShouldBeGroupedWithDataSourceStack(dataSource: DataSource, dataSourceStack: DataSourceStack) {
  return (dataSourceStack.outputName === dataSource.outputName || (!dataSourceStack.hasSourceGraphElement && !dataSource.hasCreatorPod)) &&
    arraysEqual(podsToPodAndPodCreatorStepNames(dataSource.createdPods), podsToPodAndPodCreatorStepNames(getCreatedPodsFromDataSourceStack(dataSourceStack))) &&
    (
      (!dataSourceStack.hasSourceGraphElement && !dataSource.hasCreatorPod) ||
      (
        (dataSourceStack.hasSourceGraphElement && dataSource.hasCreatorPod) &&
        (
          (dataSourceStack.sourceGraphElement.type === 'pod' && dataSourceStack.sourceGraphElement.pod.name === dataSource.creatorPod.name) ||
          (dataSourceStack.sourceGraphElement.type === 'pod-stack' && dataSourceStack.sourceGraphElement.podStack.pods.some(pod => pod.name === dataSource.creatorPod.name))
        )
      )
    );
}

export function getGraphElementIncludingPod(pod: Pod, podGraphElements: GraphElement[]) {
  for (let i = 0; i < podGraphElements.length; i++) {
    let podGraphElement: GraphElement = podGraphElements[i];
    if (podGraphElement.type === 'pod') {
      if (pod.name === podGraphElement.pod.name) {
        return podGraphElement;
      }
    }
    if (podGraphElement.type === 'pod-stack') {
      for (let j = 0; j < podGraphElement.podStack.pods.length; j++) {
        let innerPod: Pod = podGraphElement.podStack.pods[j];
        if (pod.name === innerPod.name) {
          return podGraphElement;
        }
      }
    }
  }
  return undefined;
}

export function getGraphElementIncludingDataSource(dataSource: DataSource, dataSourceGraphElements: GraphElement[]) {
  for (let i = 0; i < dataSourceGraphElements.length; i++) {
    let dataSourceGraphElement: GraphElement = dataSourceGraphElements[i];
    if (dataSourceGraphElement.type === 'data-source') {
      if (dataSource.name === dataSourceGraphElement.dataSource.name) {
        return dataSourceGraphElement;
      }
    }
    if (dataSourceGraphElement.type === 'data-source-stack') {
      for (let j = 0; j < dataSourceGraphElement.dataSourceStack.dataSources.length; j++) {
        let innerDataSource: DataSource = dataSourceGraphElement.dataSourceStack.dataSources[j];
        if (dataSource.name === innerDataSource.name) {
          return dataSourceGraphElement;
        }
      }
    }
  }
  return undefined;
}

export function getAllCurrentGraphElementsWithStacks() {
  return Array.from(currentGraphElementsWithStacksMap.keys()).map(graphElementKey => <GraphElement>currentGraphElementsWithStacksMap.get(graphElementKey)).filter(graphElement => graphElement != undefined);
}

export function setAllCurrentGraphElementsWithStacks() {
  currentGraphElementsWithStacksMap.clear();

  let podGraphElements: GraphElement[] = [];
  let currentPods: Pod[] = getCurrentPods();

  currentPods.forEach(pod => {
    for (let i = 0; i < podGraphElements.length; i++) {
      let podGraphElement = podGraphElements[i];
      if (podGraphElement.type != "pod-stack") {
        continue;
      }
      if (podShouldBeGroupedWithPodStack(pod, podGraphElement.podStack)) {
        podGraphElement.podStack.pods.push(pod);
        return;
      }
    }
    let podStackGraphElement: GraphElement = {
      type: "pod-stack",
      podStack: {
        stackId: uuidv4(),
        hasCreatorStep: pod.hasCreatorStep,
        creatorStep: pod.hasCreatorStep ? pod.creatorStep : undefined,
        pods: [pod]
      }
    };
    podGraphElements.push(podStackGraphElement);
    if (pod.hasCreatorStep) {
      pod.creatorStep.podType = 'pod-stack';
      pod.creatorStep.pods = undefined;
      pod.creatorStep.podStack = podStackGraphElement.podStack;
    }
  });

  let podGraphElementsSmallStacksUndone: GraphElement[] = podGraphElements.map(element => {
    if (element.type === 'pod-stack' && element.podStack.pods.length <= maxNumberOfSeparateGraphElements && element.podStack.pods.length > 0) {
      element.podStack.pods[0].creatorStep.podType = 'pod';
      element.podStack.pods[0].creatorStep.pods = element.podStack.pods;
      if (element.podStack.pods.length > 1) {
        for (let i = 1; i < element.podStack.pods.length; i++) {
          podGraphElements.push({type: 'pod', pod: element.podStack.pods[i]});
        }
      }
      return {type: 'pod', pod: element.podStack.pods[0]};
    } else {
      return element;
    }
  });

  let dataSourceGraphElements: GraphElement[] = [];
  let currentDataSources: DataSource[] = getCurrentDataSources();
  currentDataSources.filter(dataSource => dataSource.hasOutputName || dataSource.creatorPod == undefined).forEach(dataSource => {
    for (let i = 0; i < dataSourceGraphElements.length; i++) {
      let dataSourceGraphElement = dataSourceGraphElements[i];
      if (dataSourceGraphElement.type != "data-source-stack") {
        continue;
      }
      if (dataSourceShouldBeGroupedWithDataSourceStack(dataSource, dataSourceGraphElement.dataSourceStack)) {
        dataSourceGraphElement.dataSourceStack.dataSources.push(dataSource);
        return;
      }
    }
    dataSourceGraphElements.push({
      type: "data-source-stack",
      dataSourceStack: {
        stackId: uuidv4(),
        hasSourceGraphElement: dataSource.hasCreatorPod,
        sourceGraphElement: dataSource.hasCreatorPod ? getGraphElementIncludingPod(dataSource.creatorPod, podGraphElements) : undefined,
        outputName: dataSource.outputName,
        dataSources: [dataSource]
      }
    });
  });

  dataSourceGraphElements = dataSourceGraphElements.map(element => {
    if (element.type === 'data-source-stack' && element.dataSourceStack.dataSources.length <= maxNumberOfSeparateGraphElements && element.dataSourceStack.dataSources.length > 0) {
      if (element.dataSourceStack.dataSources.length > 1) {
        for (let i = 1; i < element.dataSourceStack.dataSources.length; i++) {
          dataSourceGraphElements.push({type: 'data-source', dataSource: element.dataSourceStack.dataSources[i]});
        }
      }
      return {type: 'data-source', dataSource: element.dataSourceStack.dataSources[0]};
    } else {
      return element;
    }
  });

  let stepGraphElements: GraphElement[] = getCurrentSteps().map(step => ({type: "step", step}));

  dataSourceGraphElements.forEach(element => {
    if (element.type === 'data-source') {
      currentGraphElementsWithStacksMap.set(element.dataSource.name, {type: 'data-source', dataSource: element.dataSource});
    }
    if (element.type === 'data-source-stack') {
      element.dataSourceStack.dataSources.forEach(dataSource => {
        dataSource.dataSourceStack = element.dataSourceStack;
      });
      currentGraphElementsWithStacksMap.set(element.dataSourceStack.stackId, {type: 'data-source-stack', dataSourceStack: element.dataSourceStack});
    }
  });

  podGraphElementsSmallStacksUndone.forEach(element => {
    if (element.type === 'pod') {
      currentGraphElementsWithStacksMap.set(element.pod.name, {type: 'pod', pod: element.pod});
    }
    if (element.type === 'pod-stack') {
      element.podStack.pods.forEach(pod => {
        pod.podStack = element.podStack;
      });
      currentGraphElementsWithStacksMap.set(element.podStack.stackId, {type: 'pod-stack', podStack: element.podStack});
    }
  });

  stepGraphElements.forEach(element => {
    if (element.type === 'step') {
      currentGraphElementsWithStacksMap.set(element.step.name, {type: 'step', step: element.step});
    }
  });
}

export function getGraphElementByIdentifier(identifier: string) {
  let graphElements: GraphElement[] = getAllCurrentGraphElementsWithStacks();
  for (let i = 0; i < graphElements.length; i++) {
    let graphElement: GraphElement = graphElements[i];
    switch (graphElement.type) {
      case "data-source":
        if (identifier === graphElement.dataSource.name) {
          return graphElement;
        }
        break;
      case "data-source-stack":
        if (identifier === graphElement.dataSourceStack.stackId) {
          return graphElement;
        }
        for (let i = 0; i < graphElement.dataSourceStack.dataSources.length; i++) {
          let dataSource: DataSource = graphElement.dataSourceStack.dataSources[i];
          if (identifier === dataSource.name) {
            return {type: 'data-source', dataSource: dataSource} as GraphElement;
          }
        }
        break;
      case "pod":
        if (identifier === graphElement.pod.name) {
          return graphElement;
        }
        break;
      case "pod-stack":
        if (identifier === graphElement.podStack.stackId) {
          return graphElement;
        }
        for (let i = 0; i < graphElement.podStack.pods.length; i++) {
          let pod: Pod = graphElement.podStack.pods[i];
          if (identifier === pod.name) {
            return {type: 'pod', pod: pod} as GraphElement;
          }
        }
        break;
      case "step":
        if (identifier === graphElement.step.name) {
          return graphElement;
        }
        break;
    }
  }
  return undefined;
}

export function getRawDataFromStep(step: Step): string {
  let completeStep = JSON.parse(step.raw);
  completeStep.spec.template = JSON.parse(step.template);
  completeStep.spec.ingest = step.ingests;
  completeStep.spec.outputs = step.outputs;
  completeStep.spec.outputs.map(output => {
    let outputLabels = {};
    output.labels.forEach(label => {
      outputLabels[label.key] = label.value;
    });
    output.labels = outputLabels;
    return output;
  });
  return JSON.stringify(completeStep);
}

export function getRawDataFromPod(pod: Pod): string {
  return pod.raw;
}

export function getRawDataFromDataSource(dataSource: DataSource): string {
  let completeDataSource = JSON.parse(dataSource.raw);
  completeDataSource.metadata.labels = {};
  dataSource.labels.forEach(label => {
    completeDataSource.metadata.labels[label.key] = label.value;
  });
  completeDataSource.spec.url = dataSource.specUrl;
  return JSON.stringify(completeDataSource);
}
