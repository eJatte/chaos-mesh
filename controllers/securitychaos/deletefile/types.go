package deletefile

import (
	"context"
	"errors"
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

	if len(securitychaos.Spec.DirectoryPath) == 0 {
		msg := "need to specify a directory path"
		err := errors.New(msg)
		e.Log.Error(err, msg)
		return err
	}

	//pvClaimName := "delete-file-pv-claim"

	err := e.CreateDummyFile(ctx)

	if err == nil {
		e.Log.Info("Created file")
	} else {
		e.Log.Error(err,"Failed to create file")
	}

	/*e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjected, "Started chaos experiment= "+" action="+string(securitychaos.Spec.Action))

	e.Log.Info("Select and filter pods")
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
			ContainerId:   containerID,
			DirectoryPath: securitychaos.Spec.DirectoryPath,
			Uid:           securitychaos.Spec.UID,
		})

		if err != nil {
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Error when deleting file")
			e.Log.Error(err, "Error when deleting file")
			return err
		}

		if response.AttackSuccessful {
			securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
			securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Deleted file. Attack succeeded.")
		} else {
			securitychaos.Status.Experiment.Message = string(v1alpha1.AttackFailedMessage)
			securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Failed to delete file. Attack failed.")
		}
	} else {
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "no pods selected")
		e.Log.Error(err, "no pods selected")
		return err
	}*/

	return nil
}

func (e *endpoint) CreateDummyFile(ctx context.Context) error {
	var user int64 = 1000
	var group int64 = 1234
	pvClaim := "delete-file-pv-claim"

	job := v12.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "create-file-job",
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
						Name:         "delete-file-pv-storage",
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvClaim,
							},
						},
					}},
					Containers: []v1.Container{
						{
							Name:  "create-file-container",
							Image: "busybox",
							Command: []string{"touch"} ,
							Args: []string{"super/data/dummyfile"} ,
							VolumeMounts: []v1.VolumeMount{{
								Name:             "delete-file-pv-storage",
								MountPath:        "/super/data",
							}},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}

	return e.Create(ctx, &job)
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
