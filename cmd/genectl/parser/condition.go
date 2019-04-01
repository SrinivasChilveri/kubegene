/*
Copyright 2019 The Kubegene Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package parser

import (
	"fmt"

	labels "k8s.io/apimachinery/pkg/labels"

	"kubegene.io/kubegene/pkg/common"
)

func validateConditionDependency(prefix string, jobName string, dependJobName string, workflow *Workflow) error {

	dependJob, ok := workflow.Jobs[dependJobName]
	if !ok {
		err := fmt.Errorf("%s: the check_result function dependecy job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	// depend job should have single command only because it should be single k8s- job related to that job

	if (len(dependJob.Commands) > 1) || (len(dependJob.CommandsIter.Vars) > 1) || (len(dependJob.CommandsIter.VarsIter) > 1) {
		err := fmt.Errorf("the check_result function dependecy job has more than one command %s dependjobName :%s", prefix, dependJobName)
		return err
	}

	currentJob, ok := workflow.Jobs[jobName]
	if !ok {
		err := fmt.Errorf("%s: the check_result function  job is missing, but the real one is %s", prefix, dependJobName)
		return err
	}

	if len(currentJob.Depends) != 1 {
		err := fmt.Errorf("%s: the check_result  job has more dependecies %v", prefix, currentJob.Depends)
		return err
	}

	for i := 0; i < len(currentJob.Depends); i++ {
		if (currentJob.Depends[i].Target == dependJobName) &&
			(currentJob.Depends[i].Type) == "whole" {
			return nil
		}
	}

	err := fmt.Errorf("%s: the check_result function dependecy job type is wrong %s", prefix, dependJobName)

	return err
}

// validateCondition validate parameter of condition is valid.
func validateCondition(jobName string, condition *ConditionInfo, inputs map[string]Input, workflow *Workflow) ErrorList {
	allErr := ErrorList{}

	if condition == nil {
		return allErr
	}
	prefix := fmt.Sprintf("workflow.%s.condition", jobName)

	for i := range condition.resultMatch {

		if IsVariant(condition.resultMatch[i].Key) {
			prefix := fmt.Sprintf("workflow.%s.condition.resultmatch.key", jobName)
			if err := ValidateVariant(prefix, condition.resultMatch[i].Key, []string{StringType}, inputs); err != nil {
				allErr = append(allErr, err)
			}
		}

		if IsVariant(condition.resultMatch[i].Operator) {
			prefix := fmt.Sprintf("workflow.%s.condition.resultmatch[%d].operator", jobName, i)
			allErr = append(allErr, fmt.Errorf("%s should not be variant", prefix))
		}
		if condition.resultMatch[i].Operator != NodeSelectorOpIn &&
			condition.resultMatch[i].Operator != NodeSelectorOpNotIn &&
			condition.resultMatch[i].Operator != NodeSelectorOpExists &&
			condition.resultMatch[i].Operator != NodeSelectorOpDoesNotExist &&
			condition.resultMatch[i].Operator != NodeSelectorOpGt &&
			condition.resultMatch[i].Operator != NodeSelectorOpLt {
			prefix := fmt.Sprintf("workflow.%s.condition.resultmatch[%d].operator", jobName, i)
			allErr = append(allErr, fmt.Errorf("%s should only be In,NotIn,DoesNotExist,Exist,Gt,Lt ", prefix))
		}
		for j := range condition.resultMatch[i].Values {
			if IsVariant(condition.resultMatch[i].Values[j]) {
				prefix := fmt.Sprintf("workflow.%s.condition.resultmatch[%d].value[%d]", jobName, i, j)
				allErr = append(allErr, fmt.Errorf("%s should not be variant", prefix))
			}
		}
	}

	if err := validateConditionDependency(prefix, jobName, condition.DependJobName, workflow); err != nil {
		allErr = append(allErr, err)
	}

	return allErr
}

func InstantiateCondition(prefix string, condition *ConditionInfo, data map[string]string) error {

	if condition == nil {
		return nil, nil
	}
	for i := range condition.resultMatch {
		if IsVariant(condition.resultMatch[i].Key) {
			condition.resultMatch[i].Key = common.ReplaceVariant(condition.resultMatch[i].Key, data)
		}
		_, err := labels.NewRequirement(condition.resultMatch[i].Key, condition.resultMatch[i].Operator,
			condition.resultMatch[i].Values)

		if err != nil {
			return err
		}
	}

}
