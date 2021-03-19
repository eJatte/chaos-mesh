package createpod

import (
	"context"
	"errors"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"net/http"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/events"

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

	var namespace = "default"

	if len(securitychaos.Spec.NameSpace) > 0 {
		namespace = securitychaos.Spec.NameSpace
	}

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "security-chaos-create-pod",
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			SecurityContext: &securitychaos.Spec.PodSecurityContext,
			Containers: []v1.Container{
				{
					Name:            "create-pod",
					Image:           "busybox",
					SecurityContext: &securitychaos.Spec.SecurityContext,
				},
			},
		},
	}

	clientSet, err := cm.InitClientSet()
	if err != nil {
		e.Log.Error(err, "Failed to init a client set")
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to init a client set")
		return err
	}

	user := securitychaos.Spec.User

	request := clientSet.KubeCli.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Body(&pod)

	if len(user) > 0 {
		request = request.SetHeader("Impersonate-User", user)
	}

	res := request.Do()

	var statusCode int
	res.StatusCode(&statusCode)
	_, err = res.Get()

	if statusCode != http.StatusCreated {
		if statusCode == http.StatusForbidden {
			securitychaos.Status.Experiment.Message = string(v1alpha1.AttackFailedMessage)
			securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)
			e.Log.Error(err, "Failed to create pod")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Failed to create pod with specified security context. Attack failed.")
		} else if err != nil {
			e.Log.Error(err, "Failed to create pod")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to create pod")
			return err
		}
	} else {
		err = e.Delete(ctx, &pod)
		if err != nil {
			e.Log.Error(err, "Failed to delete pod")
			e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to delete pod")
			return err
		}

		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Created pod with specified security context. Attack succeeded.")
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

		return chaos.Spec.Action == v1alpha1.CreatePodAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
