package deletefile

import (
	"context"
	"errors"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	"time"

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

	e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjected, "Started chaos experiment= "+" action="+string(securitychaos.Spec.Action))

	var attackSuccessful = false

	filename := "dummyfile"

	err := e.CreateDummyFile(ctx, securitychaos.Spec.UID, securitychaos.Spec.GID, securitychaos.Spec.PvClaim, filename)

	if err == nil {
		e.Log.Info("Created file")
	} else {
		e.Log.Error(err, "Failed to create file")
	}

	time.Sleep(10 * time.Second)

	pods, err := selector.SelectAndFilterPods(ctx, e.Client, e.Reader, &securitychaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		e.Log.Error(err, "failed to select and filter pods")
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "failed to select and filter pods")
		return err
	}

	if len(pods) > 0 {
		pod := pods[0]

		daemonClient, err := client.NewChaosDaemonClient(ctx, e.Client, &pod, config.ControllerCfg.ChaosDaemonPort)
		if err != nil {
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "failed to get chaos daemon client")
			e.Log.Error(err, "failed to get chaos daemon client")
			return err
		}
		defer daemonClient.Close()

		containerID := pod.Status.ContainerStatuses[0].ContainerID

		response, err := daemonClient.DeleteFile(ctx, &pb.DeleteFileRequest{
			ContainerId: containerID,
			FilePath:    securitychaos.Spec.VolumeMountPath + filename,
			Uid:         securitychaos.Spec.UID,
			Gid:         securitychaos.Spec.GID,
		})

		if err != nil {
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Error when deleting file")
			e.Log.Error(err, "Error when deleting file")
			return err
		}

		attackSuccessful = response.AttackSuccessful
	} else {
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "no pods selected")
		e.Log.Error(err, "no pods selected")
		errDelete := e.DeleteDummyFile(ctx, securitychaos.Spec.UID, securitychaos.Spec.GID, securitychaos.Spec.PvClaim, filename)
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

		err = e.DeleteDummyFile(ctx, securitychaos.Spec.UID, securitychaos.Spec.GID, securitychaos.Spec.PvClaim, filename)
		if err == nil {
			e.Log.Info("Deleted file")
		} else {
			e.Log.Error(err, "Failed to delete file")
		}
	}

	return nil
}

func (e *endpoint) CreateDummyFile(ctx context.Context, uid int64, gid int64, pvClaim string, filename string) error {
	job := e.GetJobTemplate(uid, gid, pvClaim, "create-file-job", "touch", filename)
	return e.Create(ctx, &job)
}

func (e *endpoint) DeleteDummyFile(ctx context.Context, uid int64, gid int64, pvClaim string, filename string) error {
	job := e.GetJobTemplate(uid, gid, pvClaim, "delete-file-job", "rm", filename)
	return e.Create(ctx, &job)
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
