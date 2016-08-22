// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package cloudformation

import (
	"github.com/aws/aws-sdk-go/private/waiter"
)

func (c *CloudFormation) WaitUntilStackCreateComplete(input *DescribeStacksInput) error {
	waiterCfg := waiter.Config{
		Operation:   "DescribeStacks",
		Delay:       30,
		MaxAttempts: 50,
		Acceptors: []waiter.WaitAcceptor{
			{
				State:    "success",
				Matcher:  "pathAll",
				Argument: "Stacks[].StackStatus",
				Expected: "CREATE_COMPLETE",
			},
			{
				State:    "failure",
				Matcher:  "pathAny",
				Argument: "Stacks[].StackStatus",
				Expected: "CREATE_FAILED",
			},
		},
	}

	w := waiter.Waiter{
		Client: c,
		Input:  input,
		Config: waiterCfg,
	}
	return w.Wait()
}

func (c *CloudFormation) WaitUntilStackDeleteComplete(input *DescribeStacksInput) error {
	waiterCfg := waiter.Config{
		Operation:   "DescribeStacks",
		Delay:       30,
		MaxAttempts: 25,
		Acceptors: []waiter.WaitAcceptor{
			{
				State:    "success",
				Matcher:  "pathAll",
				Argument: "Stacks[].StackStatus",
				Expected: "DELETE_COMPLETE",
			},
			{
				State:    "success",
				Matcher:  "error",
				Argument: "",
				Expected: "ValidationError",
			},
			{
				State:    "failure",
				Matcher:  "pathAny",
				Argument: "Stacks[].StackStatus",
				Expected: "DELETE_FAILED",
			},
		},
	}

	w := waiter.Waiter{
		Client: c,
		Input:  input,
		Config: waiterCfg,
	}
	return w.Wait()
}

func (c *CloudFormation) WaitUntilStackUpdateComplete(input *DescribeStacksInput) error {
	waiterCfg := waiter.Config{
		Operation:   "DescribeStacks",
		Delay:       30,
		MaxAttempts: 5,
		Acceptors: []waiter.WaitAcceptor{
			{
				State:    "success",
				Matcher:  "pathAll",
				Argument: "Stacks[].StackStatus",
				Expected: "UPDATE_COMPLETE",
			},
			{
				State:    "failure",
				Matcher:  "pathAny",
				Argument: "Stacks[].StackStatus",
				Expected: "UPDATE_FAILED",
			},
		},
	}

	w := waiter.Waiter{
		Client: c,
		Input:  input,
		Config: waiterCfg,
	}
	return w.Wait()
}
