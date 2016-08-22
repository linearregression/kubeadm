/*
Copyright 2016 The Kubernetes Authors.

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

package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/kubernetes/pkg/api"
	apierrs "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/batch"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	// How long to wait for a scheduledjob
	scheduledJobTimeout = 5 * time.Minute
)

var _ = framework.KubeDescribe("ScheduledJob", func() {
	options := framework.FrameworkOptions{
		ClientQPS:    20,
		ClientBurst:  50,
		GroupVersion: &unversioned.GroupVersion{Group: batch.GroupName, Version: "v2alpha1"},
	}
	f := framework.NewFramework("scheduledjob", options, nil)

	BeforeEach(func() {
		if _, err := f.Client.Batch().ScheduledJobs(f.Namespace.Name).List(api.ListOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				framework.Skipf("Could not find ScheduledJobs resource, skipping test: %#v", err)
			}
		}
	})

	// multiple jobs running at once
	It("should schedule multiple jobs concurrently", func() {
		By("Creating a scheduledjob")
		scheduledJob := newTestScheduledJob("concurrent", "*/1 * * * ?", batch.AllowConcurrent)
		scheduledJob, err := createScheduledJob(f.Client, f.Namespace.Name, scheduledJob)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring more than one job is running at a time")
		err = waitForActiveJobs(f.Client, f.Namespace.Name, scheduledJob.Name, 2)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring exactly two jobs exists by listing jobs explicitly")
		jobs, err := f.Client.Batch().Jobs(f.Namespace.Name).List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(jobs.Items) >= 2).To(BeTrue())

		By("Removing scheduledjob")
		err = deleteScheduledJob(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	// suspended should not schedule jobs
	It("should not schedule jobs when suspended", func() {
		By("Creating a suspended scheduledjob")
		scheduledJob := newTestScheduledJob("suspended", "*/1 * * * ?", batch.AllowConcurrent)
		scheduledJob.Spec.Suspend = newBool(true)
		scheduledJob, err := createScheduledJob(f.Client, f.Namespace.Name, scheduledJob)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring no jobs are scheduled")
		err = waitForNoJobs(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).To(HaveOccurred())

		By("Ensuring no job exists by listing jobs explicitly")
		jobs, err := f.Client.Batch().Jobs(f.Namespace.Name).List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(jobs.Items).To(HaveLen(0))

		By("Removing scheduledjob")
		err = deleteScheduledJob(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	// only single active job is allowed for ForbidConcurrent
	It("should not schedule new jobs when ForbidConcurrent", func() {
		By("Creating a ForbidConcurrent scheduledjob")
		scheduledJob := newTestScheduledJob("forbid", "*/1 * * * ?", batch.ForbidConcurrent)
		scheduledJob, err := createScheduledJob(f.Client, f.Namespace.Name, scheduledJob)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring a job is scheduled")
		err = waitForActiveJobs(f.Client, f.Namespace.Name, scheduledJob.Name, 1)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring exactly one is scheduled")
		scheduledJob, err = getScheduledJob(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(scheduledJob.Status.Active).Should(HaveLen(1))

		By("Ensuring exaclty one job exists by listing jobs explicitly")
		jobs, err := f.Client.Batch().Jobs(f.Namespace.Name).List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(jobs.Items).To(HaveLen(1))

		By("Ensuring no more jobs are scheduled")
		err = waitForActiveJobs(f.Client, f.Namespace.Name, scheduledJob.Name, 2)
		Expect(err).To(HaveOccurred())

		By("Removing scheduledjob")
		err = deleteScheduledJob(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	// only single active job is allowed for ReplaceConcurrent
	It("should replace jobs when ReplaceConcurrent", func() {
		By("Creating a ReplaceConcurrent scheduledjob")
		scheduledJob := newTestScheduledJob("replace", "*/1 * * * ?", batch.ReplaceConcurrent)
		scheduledJob, err := createScheduledJob(f.Client, f.Namespace.Name, scheduledJob)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring a job is scheduled")
		err = waitForActiveJobs(f.Client, f.Namespace.Name, scheduledJob.Name, 1)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring exactly one is scheduled")
		scheduledJob, err = getScheduledJob(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(scheduledJob.Status.Active).Should(HaveLen(1))

		By("Ensuring exaclty one job exists by listing jobs explicitly")
		jobs, err := f.Client.Batch().Jobs(f.Namespace.Name).List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(jobs.Items).To(HaveLen(1))

		By("Ensuring the job is replaced with a new one")
		err = waitForJobReplaced(f.Client, f.Namespace.Name, jobs.Items[0].Name)
		Expect(err).NotTo(HaveOccurred())

		By("Removing scheduledjob")
		err = deleteScheduledJob(f.Client, f.Namespace.Name, scheduledJob.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})

// newTestScheduledJob returns a scheduledjob which does one of several testing behaviors.
func newTestScheduledJob(name, schedule string, concurrencyPolicy batch.ConcurrencyPolicy) *batch.ScheduledJob {
	parallelism := int32(1)
	completions := int32(1)
	return &batch.ScheduledJob{
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: batch.ScheduledJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: concurrencyPolicy,
			JobTemplate: batch.JobTemplateSpec{
				Spec: batch.JobSpec{
					Parallelism: &parallelism,
					Completions: &completions,
					Template: api.PodTemplateSpec{
						Spec: api.PodSpec{
							RestartPolicy: api.RestartPolicyOnFailure,
							Volumes: []api.Volume{
								{
									Name: "data",
									VolumeSource: api.VolumeSource{
										EmptyDir: &api.EmptyDirVolumeSource{},
									},
								},
							},
							Containers: []api.Container{
								{
									Name:    "c",
									Image:   "gcr.io/google_containers/busybox:1.24",
									Command: []string{"sleep", "300"},
									VolumeMounts: []api.VolumeMount{
										{
											MountPath: "/data",
											Name:      "data",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createScheduledJob(c *client.Client, ns string, scheduledJob *batch.ScheduledJob) (*batch.ScheduledJob, error) {
	return c.Batch().ScheduledJobs(ns).Create(scheduledJob)
}

func getScheduledJob(c *client.Client, ns, name string) (*batch.ScheduledJob, error) {
	return c.Batch().ScheduledJobs(ns).Get(name)
}

func deleteScheduledJob(c *client.Client, ns, name string) error {
	return c.Batch().ScheduledJobs(ns).Delete(name, nil)
}

// Wait for at least given amount of active jobs.
func waitForActiveJobs(c *client.Client, ns, scheduledJobName string, active int) error {
	return wait.Poll(framework.Poll, scheduledJobTimeout, func() (bool, error) {
		curr, err := c.Batch().ScheduledJobs(ns).Get(scheduledJobName)
		if err != nil {
			return false, err
		}
		return len(curr.Status.Active) >= active, nil
	})
}

// Wait for no jobs to appear.
func waitForNoJobs(c *client.Client, ns, jobName string) error {
	return wait.Poll(framework.Poll, scheduledJobTimeout, func() (bool, error) {
		curr, err := c.Batch().ScheduledJobs(ns).Get(jobName)
		if err != nil {
			return false, err
		}

		return len(curr.Status.Active) != 0, nil
	})
}

// Wait for a job to be replaced with a new one.
func waitForJobReplaced(c *client.Client, ns, previousJobName string) error {
	return wait.Poll(framework.Poll, scheduledJobTimeout, func() (bool, error) {
		jobs, err := c.Batch().Jobs(ns).List(api.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(jobs.Items) != 1 {
			return false, fmt.Errorf("More than one job is running")
		}
		return jobs.Items[0].Name != previousJobName, nil
	})
}
