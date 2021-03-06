import {AfterViewInit, Component, ElementRef, Input, NgZone, ViewChild} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {DataSource, GraphElement, Pod, Step} from '../../../externalized/definitions/definitions';
import {
  getAllCurrentGraphElementsWithStacks,
  getGraphElementByIdentifier,
  getRawDataFromDataSource,
  getRawDataFromPod,
  getRawDataFromStep
} from '../../../externalized/functionalities/quality-of-life-functions';
import {ActivatedRoute, Router} from '@angular/router';
import {Location} from '@angular/common';
import {SharedService} from '../../../shared-service';
import {AbstractControl, FormArray, FormBuilder, FormControl, FormGroup, ValidatorFn, Validators} from '@angular/forms';
import {RxwebValidators} from "@rxweb/reactive-form-validators";
import {useLocalDataSources, useLocalSteps} from "../../../externalized/config/config";

function arrayContains(needle, arrhaystack) {
  return (arrhaystack.indexOf(needle) > -1);
}

const StepIngestValueRequiredValidator: ValidatorFn = (fg: FormGroup) => {
  const checkValue = fg.get('check').value;
  const valueValue = fg.get('value').value;
  if (arrayContains(checkValue, ['wildcard', 'exact', 'regex'])) {
    if (valueValue.toString().length === 0) {
      return {valueRequired: true};
    }
    return null;
  }
  return null;
};

const JsonValidator: ValidatorFn = (fc: FormControl) => {
  try {
    JSON.parse(fc.value);
  } catch (e) {
    return {jsonInvalid: true};
  }
  return null;
};

@Component({
  selector: 'app-config-modal',
  templateUrl: './config-modal.component.html',
  styleUrls: ['./config-modal.component.css']
})
export class ConfigModalComponent implements AfterViewInit {

  constructor(
    private modalService: NgbModal,
    private readonly route: ActivatedRoute,
    private readonly router: Router,
    private location: Location,
    private sharedService: SharedService,
    private fb: FormBuilder,
    private ngZone: NgZone
  ) {
  }

  @Input() currentGraphElementsWithStacksMap: Map<string, GraphElement> = new Map();

  currentGraphElement: GraphElement | undefined;
  selectedIdentifier: string | undefined;

  @ViewChild('content', {static: false}) theModal: ElementRef | undefined;

  selectedElementCache: GraphElement;

  selectedElement = () => {
    let selectedElement: GraphElement;
    if (this.selectedElementCache != undefined) {
      if (this.selectedElementCache.type === 'step') {
        if (this.selectedElementCache.step.name === this.selectedIdentifier) {
          selectedElement = this.selectedElementCache;
        }
      }
      if (this.selectedElementCache.type === 'data-source') {
        if (this.selectedElementCache.dataSource.name === this.selectedIdentifier) {
          selectedElement = this.selectedElementCache;
        }
      }
      if (this.selectedElementCache.type === 'pod') {
        if (this.selectedElementCache.pod.name === this.selectedIdentifier) {
          selectedElement = this.selectedElementCache;
        }
      }
      if (this.selectedElementCache.type === 'data-source-stack') {
        alert('selectedElementCache is DataSourceStack')
      }
      if (this.selectedElementCache.type === 'pod-stack') {
        alert('selectedElementCache is PodStack')
      }
    }

    if (selectedElement == undefined) {
      selectedElement = getGraphElementByIdentifier(this.selectedIdentifier);
    }


    selectedElement.readOnly = selectedElement.type !== 'step';
    if (selectedElement.type === 'data-source') {
      if (selectedElement.dataSource.hasCreatorPod === false) {
        selectedElement.readOnly = false;
      }
    }

    return selectedElement;
  };

  async ngAfterViewInit() {
    this.route.paramMap.subscribe(params => {
      this.ngZone.run(async () => {
        let idParam = params.get('id');
        if (idParam == null) {
          return;
        }

        await new Promise(resolve => setTimeout(() => {
          this.openModal(idParam);
          resolve();
        }, getAllCurrentGraphElementsWithStacks().length === 0 ? 2000 : 0));
      });
    });
  }

  updateUrlBySelectElement(value: string) {
    this.updateUrl('/id/' + value);
  }

  selectionChange(element: any) {
    this.updateUrlBySelectElement(element.value);
    this.fillForms();
  }

  updateUrl(url: string) {
    this.location.replaceState(url);
  }

  goto(identifier: string): void {
    this.router.navigate(['id', identifier]).then(() => {
    });
  }

  async openModal(identifier: string) {
    this.selectedIdentifier = identifier;

    this.modalService.dismissAll();
    this.currentGraphElement = getGraphElementByIdentifier(identifier);

    if (this.currentGraphElement === undefined || this.currentGraphElement === null) {
      return;
    }

    if (this.currentGraphElement.type === 'data-source-stack' && this.currentGraphElement.dataSourceStack.dataSources.length !== 0) {
      this.goto(this.currentGraphElement.dataSourceStack.dataSources[0].name);
      return;
    }
    if (this.currentGraphElement.type === 'pod-stack' && this.currentGraphElement.podStack.pods.length !== 0) {
      this.goto(this.currentGraphElement.podStack.pods[0].name);
      return;
    }

    this.fillForms();

    await this.modalService.open(this.theModal, {
      ariaLabelledBy: 'modal-basic-title',
      size: 'lg'
    }).result.then(() => {
      this.router.navigate(['']);
    }, () => {
      this.router.navigate(['']);
    });
  }

  filterGraph(graphElement: GraphElement) {
    this.sharedService.filterGraph(graphElement);
  }

  save(graphElement: GraphElement) {
    if (graphElement == undefined) {
      return;
    }

    if (graphElement.type === 'step') {
      let step = graphElement.step;

      let template: string = this.stepFormData.value['template'];
      if (template != undefined) {
        step.template = template;
      }

      step.ingests = [];
      let stepIngestsFormArray = this.stepIngestsFormArray;
      for (let i = 0; i < stepIngestsFormArray.length; i++) {
        let stepIngestsFormGroup = stepIngestsFormArray.at(i);

        let check = stepIngestsFormGroup.value['check'];
        let value = stepIngestsFormGroup.value['value'];

        if (!arrayContains(check, ['wildcard', 'exact', 'regex'])) {
          value = undefined;
        }

        let ingest = {
          key: stepIngestsFormGroup.value['key'],
          value: value,
          check: check
        };

        step.ingests.push(ingest);
      }

      step.outputs = [];
      let stepOutputsFormArray = this.stepOutputsFormArray;
      for (let i = 0; i < stepOutputsFormArray.length; i++) {
        let stepOutputsFormGroup = stepOutputsFormArray.at(i);

        let name = stepOutputsFormGroup.value['name'];
        let url = stepOutputsFormGroup.value['url'];

        let labels = [];
        let outputLabelsFormArray = stepOutputsFormGroup.get('labels') as FormArray;
        for (let j = 0; j < outputLabelsFormArray.length; j++) {
          let labelFormGroup = outputLabelsFormArray.at(j) as FormGroup;
          labels.push({
            key: labelFormGroup.value['key'],
            value: labelFormGroup.value['value']
          });
        }

        step.outputs.push({
          name: name,
          url: url,
          labels: labels
        });
      }

      if (!useLocalSteps) {
        const Http = new XMLHttpRequest();
        const url = 'http://localhost:8080/step/default';
        Http.open("POST", url);
        Http.send(getRawDataFromStep(step));

        Http.onreadystatechange = () => {
          location.reload();
        }
      }
    }
    if (graphElement.type === 'data-source') {
      let dataSource = graphElement.dataSource;

      let specUrl: string = this.dataSourceFormData.value['specUrl'];
      if (specUrl != undefined) {
        dataSource.specUrl = specUrl;
      }

      dataSource.labels = [];
      let dataSourceLabelsFormArray = this.dataSourceLabelsFormArray;
      for (let i = 0; i < dataSourceLabelsFormArray.length; i++) {
        let dataSourceLabelsFormGroup = dataSourceLabelsFormArray.at(i);
        dataSource.labels.push({
          key: dataSourceLabelsFormGroup.value['key'],
          value: dataSourceLabelsFormGroup.value['value']
        });
      }

      if (!useLocalDataSources) {
        const Http = new XMLHttpRequest();
        const url = 'http://localhost:8080/datasource/default';
        Http.open("POST", url);
        Http.send(getRawDataFromDataSource(dataSource));

        Http.onreadystatechange = () => {
          location.reload();
        }
      }
    }
    if (graphElement.type === 'pod') {
      let pod = graphElement.pod;

      let raw: string = this.podFormData.value['raw'];
      if (raw != undefined) {
        pod.raw = raw;
      }

      console.log(getRawDataFromPod(pod));
      // TODO save in kubernetes
    }
  }

  podFormData = this.fb.group({
    raw: this.fb.control('')
  });

  dataSourceFormData: FormGroup = this.fb.group({
    specUrl: this.fb.control(''),
    labels: this.fb.array([])
  });


  stepFormData = this.fb.group({
    template: this.fb.control(''),
    ingests: this.fb.array([]),
    outputs: this.fb.array([])
  });

  fillForms() {
    let graphElement = this.selectedElement();
    switch (graphElement.type) {
      case "data-source":
        this.fillDataSourceForm(graphElement.dataSource);
        break;
      case "pod":
        this.fillPodForm(graphElement.pod);
        break;
      case "step":
        this.fillStepForm(graphElement.step);
        break;
    }
  }

  handleSubmit() {
    this.save(this.selectedElement())
  }

  removeLabelFromDataSourceForm(index: number) {
    this.dataSourceLabelsFormArray.removeAt(index);
  }

  removeIngestFromStepForm(index: number) {
    this.stepIngestsFormArray.removeAt(index);
  }

  removeOutputFromStepForm(index: number) {
    this.stepOutputsFormArray.removeAt(index);
  }

  removeLabelFromStepOutputLabelsFormArray(labelsFormArray: AbstractControl, index: number) {
    (<FormArray>labelsFormArray).removeAt(index);
  }

  addLabelToDataSourceForm() {
    this.dataSourceLabelsFormArray.push(
      this.fb.group({
        key: this.fb.control('', [
          Validators.required,
          RxwebValidators.unique()
        ]),
        value: this.fb.control('', Validators.required)
      })
    );
  }

  addIngestToStepForm() {

    this.stepIngestsFormArray.push(
      this.fb.group({
        key: this.fb.control('', [
          Validators.required,
          RxwebValidators.unique()
        ]),
        value: this.fb.control(''),
        check: this.fb.control('wildcard', Validators.pattern('^(wildcard|exact|regex|present|absent)$'))
      }, {validator: StepIngestValueRequiredValidator})
    );
  }

  addOutputToStepForm() {
    this.stepOutputsFormArray.push(
      this.fb.group({
        name: this.fb.control('', [
          Validators.required,
          RxwebValidators.unique()
        ]),
        url: this.fb.control('', Validators.required),
        labels: this.fb.array([
          this.fb.group({
            key: this.fb.control('', [
              Validators.required,
              RxwebValidators.unique()
            ]),
            value: this.fb.control('', Validators.required)
          })
        ])
      })
    );
  }

  addLabelToStepOutput(stepOutputLabelsFormArray: AbstractControl) {
    (<FormArray>stepOutputLabelsFormArray).push(
      this.fb.group({
        key: this.fb.control('', [
          Validators.required,
          RxwebValidators.unique()
        ]),
        value: this.fb.control('', Validators.required)
      })
    );
  }

  get dataSourceLabelsFormArray() {
    return this.dataSourceFormData.get('labels') as FormArray;
  }

  get dataSourceSpecUrlFormControl() {
    return this.dataSourceFormData.get('specUrl') as FormControl;
  }

  get stepIngestsFormArray() {
    return this.stepFormData.get('ingests') as FormArray;
  }

  get stepOutputsFormArray() {
    return this.stepFormData.get('outputs') as FormArray;
  }

  get stepTemplateFormControl() {
    return this.stepFormData.get('template') as FormControl;
  }

  getOutputLabelsFormArray(outputGroup: AbstractControl) {
    return outputGroup.get('labels') as FormArray;
  }

  getControlFromGroup(name: string, from: AbstractControl) {
    return from.get(name) as FormControl;
  }

  getOptionsIngestsArrayFromControlForm(formControl: AbstractControl) {
    return [...[formControl.value], ...['wildcard', 'exact', 'regex', 'present', 'absent']];
  }

  ingestCheckSelectionIsValid(selection: string) {
    return arrayContains(selection, ['wildcard', 'exact', 'regex', 'present', 'absent']);
  }

  shouldDisplayIngestValue(ingestCheck: string) {
    return ingestCheck != 'present' && ingestCheck != 'absent' && this.ingestCheckSelectionIsValid(ingestCheck);
  }

  private fillDataSourceForm(dataSource: DataSource) {
    this.dataSourceFormData.setControl('specUrl', this.fb.control(dataSource.specUrl, Validators.required));
    if (this.selectedElement().readOnly) {
      this.dataSourceFormData.controls['specUrl'].disable();
    }

    let labels = this.fb.array([]);
    for (let i = 0; i < dataSource.labels?.length; i++) {
      let label = dataSource.labels[i];
      let labelGroup = this.fb.group({
        key: this.fb.control(label.key != undefined ? label.key : '', [
          Validators.required,
          RxwebValidators.unique()
        ]),
        value: this.fb.control(label.value != undefined ? label.value : '', [
          Validators.required
        ])
      });
      if (this.selectedElement().readOnly) {
        labelGroup.controls['key'].disable();
        labelGroup.controls['value'].disable();
      }
      labels.push(labelGroup);
    }
    this.dataSourceFormData.setControl('labels', labels);
  }

  private fillPodForm(pod: Pod) {
    this.podFormData.setControl('raw', this.fb.control(pod.raw));
    this.podFormData.controls['raw'].disable();
  }

  private fillStepForm(step: Step) {
    this.stepFormData.setControl('template', this.fb.control(step.template, [
      Validators.required,
      JsonValidator
    ]));

    let ingests = this.fb.array([]);
    for (let i = 0; i < step.ingests?.length; i++) {
      let ingest = step.ingests[i];

      let keyValue = () => {
        return ingest.key == undefined ? '' : ingest.key;
      };

      let valueValue = () => {
        return ingest.value == undefined ? '' : ingest.value;
      };

      let checkValue = () => {

        if (ingest.check == undefined || ingest.check === '') {
          return 'wildcard';
        }

        if (arrayContains(ingest.check, ['wildcard', 'exact', 'regex', 'present', 'absent'])) {
          return ingest.check;
        }

        return ingest.check;
      };

      let ingestGroup = this.fb.group({
        // TODO Should empty fields be '', null, or non-existent? Currently ''. Check what API provides.
        key: this.fb.control(keyValue(), [
          Validators.required,
          RxwebValidators.unique()
        ]),
        value: this.fb.control(valueValue()),
        check: this.fb.control(checkValue(), Validators.pattern('^(wildcard|exact|regex|present|absent)$'))
      }, {validator: StepIngestValueRequiredValidator});
      ingests.push(ingestGroup);
    }
    this.stepFormData.setControl('ingests', ingests);

    let outputs = this.fb.array([]);
    for (let i = 0; i < step.outputs?.length; i++) {
      let output = step.outputs[i];
      let outputGroup = this.fb.group({
        name: this.fb.control(output.name == undefined ? '' : output.name, [
          Validators.required,
          RxwebValidators.unique()
        ]),
        url: this.fb.control(output.url == undefined ? '' : output.url, Validators.required)
      });
      let outputLabels = this.fb.array([]);
      for (let j = 0; j < output.labels?.length; j++) {
        let label = output.labels[j];
        let labelGroup = this.fb.group({
          key: this.fb.control(label.key != undefined ? label.key : '', [
            Validators.required,
            RxwebValidators.unique()
          ]),
          value: this.fb.control(label.value != undefined ? label.value : '', Validators.required)
        });
        outputLabels.push(labelGroup);
      }
      outputGroup.setControl('labels', outputLabels);
      outputs.push(outputGroup);
    }
    this.stepFormData.setControl('outputs', outputs);
  }

}
