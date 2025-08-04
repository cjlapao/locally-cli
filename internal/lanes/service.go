package lanes

// import (
// 	"fmt"
// 	"strings"

// 	"github.com/cjlapao/locally-cli/internal/common"
// 	"github.com/cjlapao/locally-cli/internal/configuration"
// 	"github.com/cjlapao/locally-cli/internal/context/pipeline_component"
// 	"github.com/cjlapao/locally-cli/internal/lanes/entities"
// 	"github.com/cjlapao/locally-cli/internal/lanes/interfaces"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/bashworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/curlworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/dockerworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/dotnetworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/efmigrationsworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/gitworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/infrastructureworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/keyvaultworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/npmworker"
// 	"github.com/cjlapao/locally-cli/internal/lanes/workers/sqlworker"
// )

// var globalAutomationService *PipelineService

// type PipelineService struct {
// 	workers []interfaces.PipelineWorker
// }

// func New() *PipelineService {
// 	svc := PipelineService{
// 		workers: make([]interfaces.PipelineWorker, 0),
// 	}

// 	svc.registerWorker(sqlworker.SqlPipelineWorker{})
// 	svc.registerWorker(bashworker.BashPipelineWorker{})
// 	svc.registerWorker(infrastructureworker.InfrastructurePipelineWorker{})
// 	svc.registerWorker(gitworker.GitPipelineWorker{})
// 	svc.registerWorker(keyvaultworker.KeyvaultPipelineWorker{})
// 	svc.registerWorker(curlworker.CurlPipelineWorker{})
// 	svc.registerWorker(efmigrationsworker.EFMigrationsPipelineWorker{})
// 	svc.registerWorker(dockerworker.DockerPipelineWorker{})
// 	svc.registerWorker(dotnetworker.DotnetPipelineWorker{})
// 	svc.registerWorker(npmworker.NpmPipelineWorker{})
// 	return &svc
// }

// func Get() *PipelineService {
// 	if globalAutomationService != nil {
// 		return globalAutomationService
// 	}

// 	return New()
// }

// func (pipeline *PipelineService) registerWorker(worker interfaces.PipelineWorker) {
// 	pipeline.workers = append(pipeline.workers, worker)
// }

// func (pipeline *PipelineService) execute(task *pipeline_component.PipelineTask) error {
// 	executed := false
// 	for _, worker := range pipeline.workers {
// 		executer := worker.New()
// 		result := executer.Run(task)
// 		if result.State == entities.StateErrored {
// 			return result.Error
// 		}
// 		if result.State == entities.StateExecuted {
// 			executed = true
// 		}
// 	}

// 	if !executed {
// 		notify.Debug("Task %s was not executed", task.Name)
// 	}

// 	return nil
// }

// func (pipeline *PipelineService) validate(task *pipeline_component.PipelineTask) bool {
// 	valid := true
// 	for _, worker := range pipeline.workers {
// 		executer := worker.New()
// 		result := executer.Validate(task)
// 		if result.State != entities.StateValid && result.State != entities.StateIgnored {
// 			valid = false
// 			notify.Error(result.String())
// 		}
// 	}

// 	return valid
// }

// func (pipeline *PipelineService) GetPipelines(name string, buildDependencies bool) []*pipeline_component.Pipeline {
// 	config := configuration.Get()
// 	context := config.GetCurrentContext()
// 	result := make([]*pipeline_component.Pipeline, 0)

// 	if len(context.Pipelines) == 0 {
// 		return result
// 	}

// 	for _, pipeline := range context.Pipelines {
// 		if strings.EqualFold(pipeline.Name, name) {
// 			result = append(result, pipeline)
// 		}
// 	}

// 	return result
// }

// func (automation *PipelineService) Run(name string) error {
// 	pipelines := automation.GetPipelines(name, true)

// 	if len(pipelines) == 0 {
// 		err := fmt.Errorf("no pipelines found with name %s", name)
// 		notify.FromError(err, "Error running pipelines")
// 		return err
// 	}

// 	for _, pipeline := range pipelines {
// 		if pipeline.Disabled {
// 			notify.Info("Pipeline %s is disabled, continuing", pipeline.Name)
// 			continue
// 		}
// 		notify.Wrench("Starting to run the pipeline %s", pipeline.Name)
// 		for _, job := range pipeline.Jobs {
// 			if job.Disabled {
// 				notify.Info("Job %s for pipeline %s is disabled, continuing", job.Name, pipeline.Name)
// 				continue
// 			}
// 			if common.IsVerbose() {
// 				notify.Wrench("Starting to execute job %s for pipeline %s", job.Name, pipeline.Name)
// 			}
// 			for _, step := range job.Steps {
// 				if step.Disabled {
// 					notify.Info("Step %s in bob %s for pipeline %s is disabled, continuing", step.Name, job.Name, pipeline.Name)
// 					continue
// 				}
// 				if common.IsVerbose() {
// 					notify.Wrench("Starting to execute task %s in job %s for pipeline %s", step.Name, job.Name, pipeline.Name)
// 				}
// 				if err := automation.execute(step); err != nil {
// 					notify.FromError(err, "There was an error executing the task %s in job %s for pipeline %s", step.Name, job.Name, pipeline.Name)
// 					return err
// 				}
// 			}
// 		}

// 		if !notify.HasErrors() {
// 			notify.Success("Finished running the pipeline %s", pipeline.Name)
// 		}
// 	}

// 	return nil
// }

// func (automation *PipelineService) Validate(name string) error {
// 	pipelines := automation.GetPipelines(name, true)

// 	if len(pipelines) == 0 {
// 		err := fmt.Errorf("no pipelines found with name %s", name)
// 		notify.FromError(err, "Error running pipelines")
// 		return err
// 	}

// 	for _, pipeline := range pipelines {
// 		notify.Wrench("Starting to validate the pipeline %s", pipeline.Name)
// 		for _, job := range pipeline.Jobs {
// 			for _, step := range job.Steps {
// 				if !automation.validate(step) {
// 					err := fmt.Errorf("there was an error validating the task %s.%s.%s", pipeline.Name, job.Name, step.Name)
// 					notify.Error(err.Error())
// 					return err
// 				}
// 			}
// 		}

// 		if !notify.HasErrors() {
// 			notify.Success("Finished validating the pipeline %s", pipeline.Name)
// 		}
// 	}

// 	return nil
// }
