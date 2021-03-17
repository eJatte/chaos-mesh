package listsecrets

import (
	"context"
	"errors"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
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

	clientSet, err := cm.InitClientSet()
	if err != nil {
		e.Log.Error(err, "Failed to init a client set")
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to init a client set")
		return err
	}

	user := securitychaos.Spec.User
	namespace := securitychaos.Spec.NameSpace

	res := clientSet.KubeCli.CoreV1().RESTClient().Get().
		Resource("secrets").
		Namespace(namespace).
		SetHeader("Impersonate-User", user).
		Do()

	var statusCode int
	res.StatusCode(&statusCode)
	_, err = res.Get()

	if statusCode == http.StatusForbidden {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackFailedMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered,
			"Could not list secrets as user="+user+" in namespace="+namespace+". Attack failed.")
	} else if statusCode == http.StatusOK {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
		securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered,
			"Could list secrets as user="+user+" in namespace="+namespace+". Attack succeeded.")
	} else {
		msg := "failed when listing secrets"
		if err == nil {
			err = errors.New(msg)
		}
		e.Log.Error(err, msg)
		return err
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

		return chaos.Spec.Action == v1alpha1.ListSecretsAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
