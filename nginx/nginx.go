package nginx

import (
	"context"
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/layer5io/gokit/logger"
	"github.com/layer5io/gokit/models"
	"github.com/layer5io/meshery-nginx/internal/config"
)

// Handler provides the methods supported by the adapter
type Handler interface {
	GetName() string
	CreateInstance([]byte, string, *chan interface{}) error
	ApplyOperation(context.Context, string, string, bool) error
	ListOperations() (Operations, error)

	StreamErr(*Event, error)
	StreamInfo(*Event)
}

// handler holds the dependencies for nginx-adapter
type handler struct {
	config  config.Handler
	log     logger.Handler
	channel *chan interface{}

	kubeClient     *kubernetes.Clientset
	kubeConfigPath string
	smiChart       string
}

// New initializes email handler.
func New(c config.Handler, l logger.Handler) Handler {
	return &handler{
		config: c,
		log:    l,
	}
}

// newClient creates a new client
func (h *handler) CreateInstance(kubeconfig []byte, contextName string, ch *chan interface{}) error {

	var err error
	h.channel = ch
	h.kubeConfigPath = "/Users/abishekk/.kube/config"
	// h.kubeConfigPath, err = h.config.GetKey("kube-config-path")
	// if err != nil {
	// 	return ErrClientConfig(err)
	// }

	config, err := h.clientConfig(kubeconfig, contextName)
	if err != nil {
		return ErrClientConfig(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ErrClientSet(err)
	}

	h.kubeClient = clientset

	return nil
}

// configClient creates a config client
func (h *handler) clientConfig(kubeconfig []byte, contextName string) (*rest.Config, error) {
	if len(kubeconfig) > 0 {
		ccfg, err := clientcmd.Load(kubeconfig)
		if err != nil {
			return nil, err
		}
		if contextName != "" {
			ccfg.CurrentContext = contextName
		}
		err = writeKubeconfig(kubeconfig, contextName, h.kubeConfigPath)
		if err != nil {
			return nil, err
		}
		return clientcmd.NewDefaultClientConfig(*ccfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	}
	return rest.InClusterConfig()
}

// writeKubeconfig creates kubeconfig in local container
func writeKubeconfig(kubeconfig []byte, contextName string, path string) error {

	yamlConfig := models.Kubeconfig{}
	err := yaml.Unmarshal(kubeconfig, &yamlConfig)
	if err != nil {
		return err
	}

	yamlConfig.CurrentContext = contextName

	d, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, d, 0600)
	if err != nil {
		return err
	}

	return nil
}
