package deletefile

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
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

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjected, "Started chaos experiment= "+" action="+string(securitychaos.Spec.Action))

	e.Log.Info("Select and filter pods")
	pods, err := selector.SelectAndFilterPods(ctx, e.Client, e.Reader, &securitychaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		e.Log.Error(err, "failed to select and filter pods")
		return err
	}

	for _, pod := range pods {

		daemonClient, err := client.NewChaosDaemonClient(ctx, e.Client, &pod, config.ControllerCfg.ChaosDaemonPort)
		if err != nil {
			e.Log.Error(err, "get chaos daemon client")
			return err
		}
		defer daemonClient.Close()

		containerID := pod.Status.ContainerStatuses[0].ContainerID

		_, err = daemonClient.DeleteFile(ctx, &pb.DeleteFileRequest{
			ContainerId: containerID,
			DirectoryPath: securitychaos.Spec.DirectoryPath,
			Uid: securitychaos.Spec.UID,
		})

		if err != nil {
			e.Log.Error(err, "Delete file")
			return err
		}

		e.Log.Info("POD NAME: " + pod.Name + " containerId: " + containerID)
	}

	securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
	securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

	e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Deleted file. Attack succeeded.")

	return nil
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
