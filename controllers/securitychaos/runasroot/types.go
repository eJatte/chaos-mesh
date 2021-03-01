package runasroot

import (
	"context"
	"errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
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

	e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjected, "Started chaos experiment= "+" action="+string(securitychaos.Spec.Action))

	var user int64 = 0

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "security-chaos-run-as-root",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "run-as-root",
					Image: "busybox",
					SecurityContext: &v1.SecurityContext{
						RunAsUser: &user,
					},
				},
			},
		},
	}

	err := e.Create(ctx, &pod)
	if err != nil {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackFailedMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Failed to create a root user pod. Attack failed.")
	} else {
		err = e.Delete(ctx, &pod)
		if err != nil {
			e.Log.Error(err, "Failed to delete pod")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to delete pod")
			return err
		}

		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Created a root user pod. Attack succeeded.")
	}

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

		return chaos.Spec.Action == v1alpha1.RunAsRootAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
