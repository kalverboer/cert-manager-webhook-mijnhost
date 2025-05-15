package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"bytes"
	"os"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"

        "k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}	

	cmd.RunWebhookServer(GroupName,
		&mijnhostDNSProviderSolver{},
	)
}


var GroupName = os.Getenv("GROUP_NAME")

type mijnhostDNSProviderSolver struct {
	client *kubernetes.Clientset
}

type Config struct {
	ApplicationSecretRef corev1.SecretKeySelector `json:"applicationSecretRef"`
}


type DNSPayload struct {
        Record struct {
                Type  string `json:"type"`
                Name  string `json:"name"`
                Value *string `json:"value"`
                TTL   int    `json:"ttl"`
        } `json:"record"`
}


func (c *mijnhostDNSProviderSolver) Name() string {
	return "mijnhost"
}

func (c *mijnhostDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	
	applicationSecret, err := c.secret(cfg.ApplicationSecretRef, ch.ResourceNamespace)
	if err != nil {
		return err
	}

	domainName := ch.ResolvedZone[:len(ch.ResolvedZone)-1] 
 
        url := "https://mijn.host/api/v2/domains/"+domainName+"/dns"
        method := "PATCH"
	
        payload := DNSPayload{}
        payload.Record.Type = "TXT"
        payload.Record.Name = ch.ResolvedFQDN
        payload.Record.Value = &ch.Key
        payload.Record.TTL = 300

        jsonData, err := json.Marshal(payload)
        if err != nil {
		panic(err)
        }

        fmt.Println(payload)

        client := &http.Client {
        }
    
        req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
        if err != nil {
		fmt.Println(err)
		return err
        }

        req.Header.Add("Accept", "")
        req.Header.Add("User-Agent", "my-application/1.0.0")
        req.Header.Add("Content-Type", "application/json")
        req.Header.Add("API-Key", applicationSecret)
    
        res, err := client.Do(req)
        if err != nil {
		fmt.Println(err)
		return err
        }
        defer res.Body.Close()
    
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
		fmt.Println(err)
		return err
        }
        fmt.Println(string(body))

        return err
}

func (c *mijnhostDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	
	applicationSecret, err := c.secret(cfg.ApplicationSecretRef, ch.ResourceNamespace)
	if err != nil {
		return err
	}
        
	domainName := ch.ResolvedZone[:len(ch.ResolvedZone)-1]

        url := "https://mijn.host/api/v2/domains/"+domainName+"/dns"
        method := "PATCH"

        payload := DNSPayload{}
        payload.Record.Type = "TXT"
        payload.Record.Name = ch.ResolvedFQDN
        payload.Record.Value = nil
        payload.Record.TTL = 300

        jsonData, err := json.Marshal(payload)
        if err != nil {
		panic(err)
        }

        fmt.Println(payload)

        client := &http.Client {
        }

        req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
        if err != nil {
		fmt.Println(err)
		return err
        }

        req.Header.Add("Accept", "")
        req.Header.Add("User-Agent", "my-application/1.0.0")
        req.Header.Add("Content-Type", "application/json")
        req.Header.Add("API-Key", applicationSecret)

        res, err := client.Do(req)
        if err != nil {
		fmt.Println(err)
		return err
        }
        defer res.Body.Close()

        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
		fmt.Println(err)
		return err
        }
        fmt.Println(string(body))

        return err
}

func (c *mijnhostDNSProviderSolver) Initialize(cfg *rest.Config, stopCh <-chan struct{}) error {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	c.client = client
	return nil
}

func (s *mijnhostDNSProviderSolver) secret(ref corev1.SecretKeySelector, namespace string) (string, error) {
	if ref.Name == "" {
		return "", nil
	}

	secret, err := s.client.CoreV1().Secrets(namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}


	bytes, ok := secret.Data[ref.Key]
	if !ok {
		return "", fmt.Errorf("key not found %q in secret '%s/%s'", ref.Key, namespace, ref.Name)
	}
	return string(bytes), err
}

func loadConfig(cfgJSON *extapi.JSON) (Config, error) {
	cfg := Config{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding config: %v", err)
	}

	return cfg, nil
}
