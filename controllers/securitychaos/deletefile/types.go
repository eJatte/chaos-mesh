package deletefile

import (
	"context"
	"errors"
	"fmt"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"strings"
	"time"

	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"

	v12 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {

	securitychaos, ok := chaos.(*v1alpha1.SecurityChaos)
	if !ok {
		err := errors.New("chaos is not SecurityChaos")
		e.Log.Error(err, "chaos is not SecurityChaos", "chaos", chaos)
		return err
	}

	if len(securitychaos.Spec.PvClaim) == 0 {
		msg := "need to specify a directory path"
		err := errors.New(msg)
		e.Log.Error(err, msg)
		return err
	}

	var uid int64 = 1000
	var gid int64 = 1000

	if securitychaos.Spec.UID > 0 {
		uid = securitychaos.Spec.UID
	}
	if securitychaos.Spec.GID > 0 {
		gid = securitychaos.Spec.GID
	}

	e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjected, "Started chaos experiment= "+" action="+string(securitychaos.Spec.Action))

	var attackSuccessful = false

	filename := "dummyfile"

	err := e.CreateDummyFile(ctx, uid, gid, securitychaos.Spec.PvClaim, filename)

	if err == nil {
		e.Log.Info("Created file")
	} else {
		e.Log.Error(err, "Failed to create file")
		return err
	}

	pods, err := selector.SelectAndFilterPods(ctx, e.Client, e.Reader, &securitychaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		e.Log.Error(err, "failed to select and filter pods")
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "failed to select and filter pods")
		return err
	}

	if len(pods) > 0 {
		pod := pods[0]

		volumeName, err := GetVolumeNameFromPvClaim(pod, securitychaos.Spec.PvClaim)
		if err != nil {
			e.Log.Error(err, "Pod does not have specified pv claim")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Pod does not have specified pv claim")
			return err
		}

		mountPath, err := GetContainerMountPathFromPvClaim(pod.Spec.Containers[0], volumeName)
		if err != nil {
			e.Log.Error(err, "Container has not mounted volume")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Container has not mounted volume")
			return err
		}

		if !strings.HasSuffix(mountPath, "/") {
			mountPath = mountPath + "/"
		}

		clientSet, err := cm.InitClientSet()
		if err != nil {
			e.Log.Error(err, "Failed to init a client set")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to init a client set")
			return err
		}

		cmd := fmt.Sprintf("rm -rf %s", mountPath + filename)

		_, err = cm.Exec(ctx, pod, cmd, clientSet.KubeCli)
		if err != nil {
			e.Log.Info("Failed to remove file from container")
			attackSuccessful = false
		} else {
			attackSuccessful = true
		}
	} else {
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "no pods selected")
		e.Log.Error(err, "no pods selected")
		errDelete := e.DeleteDummyFile(ctx, uid, gid, securitychaos.Spec.PvClaim, filename)
		if errDelete == nil {
			e.Log.Info("Deleted file")
		} else {
			e.Log.Error(errDelete, "Failed to delete file")
		}
		return err
	}

	if attackSuccessful {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Deleted file. Attack succeeded.")
	} else {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackFailedMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Failed to delete file. Attack failed.")

		err = e.DeleteDummyFile(ctx, uid, securitychaos.Spec.GID, securitychaos.Spec.PvClaim, filename)
		if err == nil {
			e.Log.Info("Deleted file")
		} else {
			e.Log.Error(err, "Failed to delete file")
			return err
		}
	}

	return nil
}

func GetContainerMountPathFromPvClaim(container v1.Container, volumeName string) (string, error) {
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == volumeName {
			return volumeMount.MountPath, nil
		}
	}
	msg := "Container has not mounted volume"
	err := errors.New(msg)

	return "", err
}

func GetVolumeNameFromPvClaim(pod v1.Pod, pvClaimName string) (string, error) {
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim.ClaimName == pvClaimName {
			return volume.Name, nil
		}
	}

	msg := "Pod does not have volume with specified claim"
	err := errors.New(msg)

	return "", err
}

func (e *endpoint) CreateDummyFile(ctx context.Context, uid int64, gid int64, pvClaim string, filename string) error {
	job := e.GetJobTemplate(uid, gid, pvClaim, "create-file-job", "touch", filename)
	return e.RunJob(ctx, job)
}

func (e *endpoint) DeleteDummyFile(ctx context.Context, uid int64, gid int64, pvClaim string, filename string) error {
	job := e.GetJobTemplate(uid, gid, pvClaim, "delete-file-job", "rm", filename)
	return e.RunJob(ctx, job)
}

type PropagationOptions struct{}

func (p PropagationOptions) ApplyToDelete(options *client2.DeleteOptions) {
	var prop = metav1.DeletePropagationForeground
	options.PropagationPolicy = &prop
}

func (e *endpoint) RunJob(ctx context.Context, job v12.Job) error {
	err := e.Create(ctx, &job)
	if err != nil {
		return err
	}

	errWait := e.WaitForJob(ctx, job)

	var p PropagationOptions

	err = e.Delete(ctx, &job, p)
	if err != nil {
		return err
	}
	if errWait != nil {
		return errWait
	}

	return nil
}

func (e *endpoint) isJobComplete(ctx context.Context, jobName string, namespace string) (bool, error) {
	var job v12.Job

	e.Log.Info("Checking if Job with name: " + jobName + " in namespace: " + namespace + " is complete")

	err := e.Reader.Get(ctx, client2.ObjectKey{
		Namespace: namespace,
		Name:      jobName,
	}, &job)

	if err != nil {
		e.Log.Error(err, "failed to get job")
		return false, err
	}

	for _, condition := range job.Status.Conditions {
		if condition.Type == v12.JobComplete {
			return true, nil
		}
	}

	return false, nil
}

func (e *endpoint) WaitForJob(ctx context.Context, job v12.Job) error {
	var completed = false
	var jobCompleteErr error = nil

	for !completed {
		time.Sleep(500 * time.Millisecond)
		completed, jobCompleteErr = e.isJobComplete(ctx, job.Name, job.Namespace)
		if jobCompleteErr != nil {
			completed = true
		}
	}
	return jobCompleteErr
}

func (e *endpoint) GetJobTemplate(user int64, group int64, pvClaim string, jobName string, command string, filename string) v12.Job {
	mountpath := "/dummyfolder/data"

	return v12.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "default",
		},
		Spec: v12.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser:  &user,
						RunAsGroup: &group,
						FSGroup:    &group,
					},
					Volumes: []v1.Volume{{
						Name: "pv-storage",
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvClaim,
							},
						},
					}},
					Containers: []v1.Container{
						{
							Name:    "create-file-container",
							Image:   "busybox",
							Command: []string{command},
							Args:    []string{mountpath + "/" + filename},
							VolumeMounts: []v1.VolumeMount{{
								Name:      "pv-storage",
								MountPath: mountpath,
							}},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.SecurityChaos{}
}

func init() {
	router.Register("securitychaos", &v1alpha1.SecurityChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.SecurityChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.DeleteFileAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
