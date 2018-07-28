package apiserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/ryutah/kubernetes-transcribe/pkg/health"
)

// TODO: this basic interface is duplicated in N places.  consolidate?
type httpGet interface {
	Get(url string) (*http.Response, error)
}

type server struct {
	addr string
	port int
}

// validator is responsible for validating the cluster and serving
type validator struct {
	servers map[string]server
	client  httpGet
}

func (s *server) check(client httpGet) (health.Status, string, error) {
	resp, err := client.Get("http://" + net.JoinHostPort(s.addr, strconv.Itoa(s.port)) + "/healthz")
	if err != nil {
		return health.Unknown, "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return health.Unknown, string(data), err
	}
	if resp.StatusCode != http.StatusOK {
		return health.Unhealthy, string(data),
			fmt.Errorf("unhealthy http status code: %d (%s)", resp.StatusCode, resp.Status)
	}
	return health.Healthy, string(data), nil
}

type ServerStatus struct {
	Component  string        `json:"component,omitempty"`
	Health     string        `json:"health,omitempty"`
	HealthCode health.Status `json:"healthCode,omitempty"`
	Msg        string        `json:"msg,omitempty"`
	Err        string        `json:"err,omitempty"`
}

func (v *validator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reply := []ServerStatus{}
	for name, server := range v.servers {
		status, msg, err := server.check(v.client)
		var errorMsg string
		if err != nil {
			errorMsg = err.Error()
		} else {
			errorMsg = "nil"
		}
		reply = append(reply, ServerStatus{
			Component:  name,
			Health:     status.String(),
			HealthCode: status,
			Msg:        msg,
			Err:        errorMsg,
		})
	}
	data, err := json.Marshal(reply)
	log.Printf("FOO: %s", string(data))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// NewValidator creates a validator for a set of servers.
func NewValidator(servers map[string]string) (http.Handler, error) {
	result := map[string]server{}
	for name, value := range servers {
		host, port, err := net.SplitHostPort(value)
		if err != nil {
			return nil, fmt.Errorf("invalid server spec: %s (%v)", value, err)
		}
		val, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid server spec: %s (%v)", port, err)
		}
		result[name] = server{addr: host, port: val}
	}
	return &validator{
		servers: result,
		client:  &http.Client{},
	}, nil
}

func makeTestValidator(servers map[string]string, get httpGet) (http.Handler, error) {
	v, e := NewValidator(servers)
	if e == nil {
		v.(*validator).client = get
	}
	return v, e
}
