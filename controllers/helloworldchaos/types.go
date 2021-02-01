package helloworldchaos

import (
	"context"
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
	e.Log.Info("Hello World!")
	e.Log.Info("Ostpojke")
	var p int32 = 3
	var user int64 = 0

	var labelMap map[string]string
	labelMap = make(map[string]string)
	labelMap["app"] = "hello-kubernetes"

	var labelSelector = metav1.LabelSelector{
		MatchLabels:      labelMap,
		MatchExpressions: nil,
	}

	var selectorMap, err = metav1.LabelSelectorAsMap(&labelSelector)

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
	}
	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

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
