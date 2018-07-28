package apiserver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/golang/glog"
)

// TODO: replace with proxy handler on minions
func handleProxyMinion(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimLeft(r.URL.Path, "/")
	rawQuery := r.URL.RawQuery

	// Expect path as: ${minion}/${query_to_minion}
	// and query_to_minion can be any query that kubelet will accept.
	//
	// For example:
	// To query stats of a minion or pod or a container,
	// path string can be ${minion}/stats/<podid>/<containerName> or
	// ${minion}/podInfo?podID=<podid>
	//
	// To query logs on a minion, path string can be:
	// ${minion}/logs/
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		badGatewayError(w, r)
		return
	}
	minionHost := parts[0]
	_, port, _ := net.SplitHostPort(minionHost)
	if port == "" {
		// Couldn't retrieve port information
		// TODO: Retrieve port info from a common object
		minionHost += ":10250"
	}
	minionPath := "/" + parts[1]

	minionURL := &url.URL{
		Scheme: "http",
		Host:   minionHost,
	}
	newReq, err := http.NewRequest("GET", minionPath+"?"+rawQuery, nil)
	if err != nil {
		glog.Errorf("Failed to create request: %s", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(minionURL)
	proxy.Transport = &minionTransport{}
	proxy.ServeHTTP(w, newReq)
}

type minionTransport struct{}

func (t *minionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			message := fmt.Sprintf("Failed to connect to minion:%s", req.URL.Host)
			resp := &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       ioutil.NopCloser(strings.NewReader(message)),
			}
			return resp, nil
		}
		return nil, err
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/plain") {
		// Do nothing, simply pass through
		return resp, err
	}

	resp, err = t.ProcessResponse(req, resp)
	return resp, err
}

func (t *minionTransport) ProcessResponse(req *http.Request, resp *http.Response) (*http.Response, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// copying the response body did not work
		return nil, err
	}

	bodyNode := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}
	nodes, err := html.ParseFragment(bytes.NewBuffer(body), bodyNode)
	if err != nil {
		glog.Errorf("Failed to found <body> node: %v", err)
		return resp, err
	}

	// Define the method to traverse the doc tree and update href node to
	// point to correct minion
	var updateHRef func(*html.Node)
	updateHRef = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for i, attr := range n.Attr {
				if attr.Key == "href" {
					URL := &url.URL{
						Path: "/proxy/minion/" + req.URL.Host + req.URL.Path + attr.Val,
					}
					n.Attr[i].Val = URL.String()
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			updateHRef(c)
		}
	}

	newContent := &bytes.Buffer{}
	for _, n := range nodes {
		updateHRef(n)
		err = html.Render(newContent, n)
		if err != nil {
			glog.Errorf("Failed to render: %v", err)
		}
	}

	resp.Body = ioutil.NopCloser(newContent)
	// Update header node with new content-length
	// TODO: Remove any hash/signature headers here?
	resp.Header.Del("Content-Length")
	resp.ContentLength = int64(newContent.Len())

	return resp, err
}
