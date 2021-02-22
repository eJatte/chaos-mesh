package helloworldchaos

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"

	"errors"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {

	e.Log.Info("Apply helloworld chaos")
	helloworldchaos, ok := chaos.(*v1alpha1.HelloWorldChaos)
	if !ok {
		return errors.New("chaos is not helloworldchaos")
	}

	e.Log.Info("Select and filter pods")
	pods, err := selector.SelectAndFilterPods(ctx, e.Client, e.Reader, &helloworldchaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		e.Log.Error(err, "failed to select and filter pods")
		return err
	}

	for _, pod := range pods {
		e.Log.Info("New chaos daemon client")
		daemonClient, err := client.NewChaosDaemonClient(ctx, e.Client, &pod, config.ControllerCfg.ChaosDaemonPort)
		if err != nil {
			e.Log.Error(err, "get chaos daemon client")
			return err
		}
		defer daemonClient.Close()
		if len(pod.Status.ContainerStatuses) == 0 {
			return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
		}

		containerID := pod.Status.ContainerStatuses[0].ContainerID
		e.Log.Info("Exec hello world chaos")
		_, err = daemonClient.ExecHelloWorldChaos(ctx, &pb.ExecHelloWorldRequest{
			ContainerId: containerID,
		})
		if err != nil {
			e.Log.Error(err, "failed to exec hello world chaos")
			return err
		}
	}

	/*var p int32 = 3
	var user int64 = 0

	var labelMap map[string]string
	labelMap = make(map[string]string)
	labelMap["app"] = "hello-kubernetes"

	var labelSelector = metav1.LabelSelector{
		MatchLabels:      labelMap,
		MatchExpressions: nil,
	}

	selectorMap, err := metav1.LabelSelectorAsMap(&labelSelector)

	if err != nil {
		e.Log.Error(err, "Failed to create map from label selector")
		return err
	}

	replicationController := v1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-kubernetes",
			Namespace: "chaos-testing",
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &p,
			Selector: selectorMap,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelMap,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "hello-kubernetes",
							Image: "paulbouwer/hello-kubernetes:1.8",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							SecurityContext: &v1.SecurityContext{
								RunAsUser: &user,
							},
						},
					},
				},
			},
		},
	}
	err = e.Create(ctx, &replicationController)
	if err != nil {
		e.Log.Error(err, "Failed to create replication controller")
		return err
	}*/
	return nil
}

// Recover means the reconciler recovers the chaos action
func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

// Object would return the instance of chaos
func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.HelloWorldChaos{}
}

func init() {
	router.Register("helloworldchaos", &v1alpha1.HelloWorldChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
