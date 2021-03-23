package kubeletapi

import (
	"context"
	"crypto/tls"
	"errors"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	securitychaos.Status.Experiment.Action = string(securitychaos.Spec.Action)

	e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjected, "Started chaos experiment= "+" action="+string(securitychaos.Spec.Action))

	clientSet, err := cm.InitClientSet()
	if err != nil {
		e.Log.Error(err, "Failed to init a client set")
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, "Failed to init a client set")
		return err
	}

	nodeList, err := clientSet.KubeCli.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		msg := "failed to list nodes"
		e.Log.Error(err, msg)
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, msg)
		return err
	}

	var host string

	for _, node := range nodeList.Items {
		if node.Name == securitychaos.Spec.Node {
			host = GetHostnameAddress(node.Status.Addresses)
			break
		}
	}

	if len(host) == 0 {
		msg := "could not find note with name "+securitychaos.Spec.Node
		err := errors.New(msg)
		e.Log.Error(err, msg)
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosInjectFailed, msg)
		return err
	}

	resp, err := MakeRequestToKubeletAPI(host)

	if err != nil {
		e.Log.Error(err, "failed to make request")
		return err
	}

	if resp.StatusCode == http.StatusOK {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackSucceededMessage)
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Successfully made request to Kubelet API. Attack succeeded.")
	} else if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		securitychaos.Status.Experiment.Message = string(v1alpha1.AttackFailedMessage)
		e.Event(securitychaos, v1.EventTypeNormal, events.ChaosRecovered, "Failed to make request to Kubelet API. Attack failed.")
	}

	return nil
}

func MakeRequestToKubeletAPI(host string) (*http.Response, error) {
	request, err := http.NewRequest("GET", "https://"+host+":10250/pods", nil)

	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func GetHostnameAddress(addresses []v1.NodeAddress) string {
	var hostname string
	var internal string
	var external string

	for _, address := range addresses {
		if address.Type == v1.NodeHostName {
			hostname = address.Address
		} else if address.Type == v1.NodeInternalIP {
			internal = address.Address
		} else if address.Type == v1.NodeExternalIP {
			external = address.Address
		}
	}

	if len(hostname) > 0 {
		return hostname
	}
	if len(internal) > 0 {
		return internal
	}

	return external
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

		return chaos.Spec.Action == v1alpha1.KubeletAPIAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
