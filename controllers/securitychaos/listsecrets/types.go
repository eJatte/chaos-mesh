package listsecrets

import (
	"context"
	"errors"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"net/http"
	"strconv"

	v1 "k8s.io/api/core/v1"
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

	clientSet, err := cm.InitClientSet()
	if err != nil {
		e.Log.Error(err, "Failed to init a client set")
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to init a client set")
		return err
	}

	secretReq := clientSet.KubeCli.CoreV1().RESTClient().Get().
		Resource("secrets").
		Namespace("default").
		SetHeader("Impersonate-User", "orion")

	e.Log.Info("Attempting to execute request with URL: " + secretReq.URL().String())

	res := secretReq.Do()

	var statusCode int

	res.StatusCode(&statusCode)

	resObject, err := res.Get()

	e.Log.Info("Status: " + strconv.Itoa(statusCode))

	if statusCode != http.StatusOK {
		e.Log.Info("COULD NOT GET SECRETS")
		if err != nil {
			e.Log.Error(err, "Error when getting secrets")
		}
	} else {
		e.Log.Info("COULD GET SECRETS")
		var secrets = resObject.(*v1.SecretList)

		for _, secret := range secrets.Items {
			e.Log.Info("SECRET: " + secret.Name)
		}
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
