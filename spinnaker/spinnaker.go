// Copyright 2016 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package spinnaker provides an interface to the Spinnaker API
package spinnaker

import (
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/pkcs12"

	"github.com/pkg/errors"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/config"
	D "github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/deps"
)

// Spinnaker implements the deploy.Deployment interface by querying Spinnaker
// and the chaosmonkey.Termination interface by terminating via Spinnaker API
// calls
type Spinnaker struct {
	endpoint string
	client   *http.Client
	user     string
}

// spinnakerClusters maps account name (e.g., "prod", "test") to a list
// of cluster names
type spinnakerClusters map[string][]string

// spinnakerServerGroup represents an autoscaling group, also called a server group,
// as represented by Spinnaker API
type spinnakerServerGroup struct {
	Name      string
	Region    string
	Disabled  bool
	Instances []spinnakerInstance
}

// spinnakerInstance represents an instance as represented by Spinnaker API
type spinnakerInstance struct {
	Name string
}

// getClient takes PKCS#12 data (encrypted cert data in .p12 format) and the
// password for the encrypted cert, and returns an http client that does TLS client auth
func getClient(pfxData []byte, password string) (*http.Client, error) {
	blocks, err := pkcs12.ToPEM(pfxData, password)
	if err != nil {
		return nil, errors.Wrap(err, "pkcs.ToPEM failed")
	}

	// The first block is the cert and the last block is the private key
	certPEMBlock := pem.EncodeToMemory(blocks[0])
	keyPEMBlock := pem.EncodeToMemory(blocks[len(blocks)-1])

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, errors.Wrap(err, "tls.X509KeyPair failed")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// getClientX509 takes X509 data (Public and Private keys) and the
// and returns an http client that does TLS client auth
func getClientX509(x509Cert, x509Key string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(x509Cert, x509Key)
	if err != nil {
		return nil, errors.Wrap(err, "tls.X509KeyPair failed")
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}

// NewFromConfig returns a Spinnaker based on config
func NewFromConfig(cfg *config.Monkey) (Spinnaker, error) {
	spinnakerEndpoint := cfg.SpinnakerEndpoint()
	certPath := cfg.SpinnakerCertificate()
	encryptedPassword := cfg.SpinnakerEncryptedPassword()
	user := cfg.SpinnakerUser()
	x509Cert := cfg.SpinnakerX509Cert()
	x509Key := cfg.SpinnakerX509Key()

	if spinnakerEndpoint == "" {
		return Spinnaker{}, errors.New("FATAL: no spinnaker endpoint specified in config")
	}

	var password string
	var err error
	var decryptor chaosmonkey.Decryptor

	if encryptedPassword != "" {
		decryptor, err = deps.GetDecryptor(cfg)
		if err != nil {
			return Spinnaker{}, err
		}

		password, err = decryptor.Decrypt(encryptedPassword)
		if err != nil {
			return Spinnaker{}, err
		}
	}

	return New(spinnakerEndpoint, certPath, password, x509Cert, x509Key, user)

}

// New returns a Spinnaker using a .p12 cert at certPath encrypted with
// password or x509 cert. The user argument identifies the email address of the user which is
// sent in the payload of the terminateInstances task API call
func New(endpoint string, certPath string, password string, x509Cert string, x509Key string, user string) (Spinnaker, error) {
	var client *http.Client
	var err error

	if x509Cert != "" && certPath != "" {
		return Spinnaker{}, errors.New("cannot use both p12 and x509 certs, choose one")
	}

	if certPath != "" {
		pfxData, err := ioutil.ReadFile(certPath)
		if err != nil {
			return Spinnaker{}, errors.Wrapf(err, "failed to read file %s", certPath)
		}

		client, err = getClient(pfxData, password)
		if err != nil {
			return Spinnaker{}, err
		}
	} else if x509Cert != "" {
		client, err = getClientX509(x509Cert, x509Key)
		if err != nil {
			return Spinnaker{}, err
		}
	} else {
		client = new(http.Client)
	}

	return Spinnaker{endpoint: endpoint, client: client, user: user}, nil
}

// AccountID returns numerical ID associated with an AWS account
func (s Spinnaker) AccountID(name string) (id string, err error) {
	url := s.accountURL(name)

	resp, err := s.client.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "could not retrieve account info for %s from spinnaker url %s", name, url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "failed to close response body from %s", url)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read body from url %s", url)
	}

	var info struct {
		AccountID string `json:"accountId"`
		Error     string `json:"error"`
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return "", errors.Wrapf(err, "could not parse body of %s as json, body: %s, error", url, body)
	}

	if resp.StatusCode != http.StatusOK {
		if info.Error == "" {
			return "", errors.Errorf("%s returned unexpected status code: %d, body: %s", url, resp.StatusCode, body)
		}

		return "", errors.New(info.Error)
	}

	// Some backends may not have associated account ids
	if info.AccountID == "" {
		return s.alternateAccountID(name)
	}

	return info.AccountID, nil

}

// alternateAccountID returns an account ID for accounts that don't have their
// own ids.
func (s Spinnaker) alternateAccountID(name string) (string, error) {

	// Sanity check: this should never be called with "prod" or "test" as an
	// argument, since this would result in infinite recursion
	if name == "prod" || name == "test" {
		return "", fmt.Errorf("alternateAccountID called with forbidden arg: %s", name)
	}

	// Heuristic: if account name has "test" in the name, we return the "test"
	// account id, otherwise with  we use the "prod" account id
	if strings.Contains(name, "test") {
		return s.AccountID("test")
	}

	return s.AccountID("prod")
}

// Apps implements deploy.Deployment.Apps
func (s Spinnaker) Apps(c chan<- *D.App, appNames []string) {
	// Close the channel we're done
	defer close(c)

	for _, appName := range appNames {
		app, err := s.GetApp(appName)
		if err != nil {
			// If we have a problem with one app, we go to the next one
			log.Printf("WARNING: GetApp failed for %s: %v", appName, err)
			continue
		}

		c <- app
	}
}

// GetInstanceIDs gets the instance ids for a cluster
func (s Spinnaker) GetInstanceIDs(app string, account D.AccountName, cloudProvider string, region D.RegionName, cluster D.ClusterName) (D.ASGName, []D.InstanceID, error) {
	url := s.activeASGURL(app, string(account), string(cluster), cloudProvider, string(region))

	resp, err := s.client.Get(url)
	if err != nil {
		return "", nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var data struct {
		Name      string
		Instances []struct{ Name string }
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to parse json at %s", url)
	}

	asg := D.ASGName(data.Name)
	instances := make([]D.InstanceID, len(data.Instances))
	for i, instance := range data.Instances {
		instances[i] = D.InstanceID(instance.Name)
	}

	return asg, instances, nil

}

// GetApp implements deploy.Deployment.GetApp
func (s Spinnaker) GetApp(appName string) (*D.App, error) {
	// data arg is a map like {accountName: {clusterName: {regionName: {asgName: [instanceId]}}}}
	data := make(D.AppMap)
	for account, clusters := range s.clusters(appName) {
		cloudProvider, err := s.CloudProvider(account)
		if err != nil {
			return nil, errors.Wrap(err, "retrieve cloud provider failed")
		}
		account := D.AccountName(account)
		data[account] = D.AccountInfo{
			CloudProvider: cloudProvider,
			Clusters:      make(map[D.ClusterName]map[D.RegionName]map[D.ASGName][]D.InstanceID),
		}
		for _, clusterName := range clusters {
			clusterName := D.ClusterName(clusterName)
			data[account].Clusters[clusterName] = make(map[D.RegionName]map[D.ASGName][]D.InstanceID)
			asgs, err := s.asgs(appName, string(account), string(clusterName))
			if err != nil {
				log.Printf("WARNING: could not retrieve asgs for app:%s account:%s cluster:%s : %v", appName, account, clusterName, err)
				continue
			}
			for _, asg := range asgs {

				// We don't terminate instances in disabled ASGs
				if asg.Disabled {
					continue
				}

				region := D.RegionName(asg.Region)
				asgName := D.ASGName(asg.Name)

				_, present := data[account].Clusters[clusterName][region]
				if !present {
					data[account].Clusters[clusterName][region] = make(map[D.ASGName][]D.InstanceID)
				}

				data[account].Clusters[clusterName][region][asgName] = make([]D.InstanceID, len(asg.Instances))

				for i, instance := range asg.Instances {
					data[account].Clusters[clusterName][region][asgName][i] = D.InstanceID(instance.Name)
				}
			}
		}
	}
	return D.NewApp(appName, data), nil
}

// AppNames returns list of names of all apps
func (s Spinnaker) AppNames() (appnames []string, err error) {
	url := s.appsURL()
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of apps from spinnaker url %s: %v", url, err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body from %s: %v", url, err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body when retrieving spinnaker app names from %s: %v", url, err)
	}
	var apps []spinnakerApp
	err = json.Unmarshal(body, &apps)
	if err != nil {
		return nil, fmt.Errorf("could not parse spinnaker apps list from %s: body: \"%s\": %v", url, string(body), err)
	}

	result := make([]string, len(apps))
	for i, app := range apps {
		result[i] = app.Name
	}

	return result, nil

}

// spinnakerApp returns an app as represented by the Spinnaker API
type spinnakerApp struct {
	Name string
}

// clusters returns a map from account name to list of cluster names
func (s Spinnaker) clusters(appName string) spinnakerClusters {
	url := s.clustersURL(appName)
	resp, err := s.client.Get(url)
	if err != nil {
		log.Println("Error connecting to spinnaker clusters endpoint")
		log.Println(url)
		log.Fatalln(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body of %s: %v", url, err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error retrieving spinnaker clusters for app", appName)
		log.Println(url)
		log.Println(string(body))
		log.Fatalln(err)
	}

	// Example cluster output:
	/*
		{
		  "prod": [
			"abc-prod"
		  ],
		  "test": [
			"abc-beta"
		  ]
		}
	*/
	var m spinnakerClusters

	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Println("Error parsing body when retrieving cluster info for", appName)
		log.Println(url)
		log.Println(string(body))
		log.Fatalln(err)
	}

	return m
}

// asgs returns a slice of autoscaling groups associated with the given cluster
func (s Spinnaker) asgs(appName, account, clusterName string) (result []spinnakerServerGroup, err error) {
	url := s.serverGroupsURL(appName, account, clusterName)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server groups url (%s): %v", url, err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close response body of %s: %v", url, err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body of server groups url (%s): body: '%s': %v", url, string(body), err)
	}

	// Example:
	/*
		[
		  {
		    "name": "abc-prod-v016",
		    "region": "us-east-1",
		    "zones": [
		      "us-east-1c",
		      "us-east-1d",
		      "us-east-1e"
		    ],
		    "disabled": false,
		    "instances": [
		      {
		        "name": "i-f9ffb752",
				...
			  },
			...
		   ]
		  }
		]
	*/

	var asgs []spinnakerServerGroup
	err = json.Unmarshal(body, &asgs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body of spinnaker asgs url (%s): body: '%s'. %v", url, string(body), err)
	}

	return asgs, nil
}

// CloudProvider returns the cloud provider for a given account
func (s Spinnaker) CloudProvider(account string) (provider string, err error) {
	url := s.accountURL(account)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("http get failed at %s", url))
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrap(err, fmt.Sprintf("body close failed at %s", url))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var fields struct {
		CloudProvider string `json:"cloudProvider"`
		Error         string `json:"error"`
	}

	err = json.Unmarshal(body, &fields)
	if err != nil {
		return "", errors.Wrap(err, "json unmarshal failed")
	}

	if resp.StatusCode != http.StatusOK {
		if fields.Error == "" {
			return "", fmt.Errorf("unexpected status code: %d. body: %s", resp.StatusCode, body)
		}

		return "", fmt.Errorf("unexpected status code: %d. error: %s", resp.StatusCode, fields.Error)
	}

	if fields.CloudProvider == "" {
		return "", fmt.Errorf("no cloudProvider field in response body")
	}

	return fields.CloudProvider, nil
}

// GetClusterNames returns a list of cluster names for an app
func (s Spinnaker) GetClusterNames(app string, account D.AccountName) (clusters []D.ClusterName, err error) {
	url := s.appURL(app)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var pcl struct {
		Clusters map[D.AccountName][]struct {
			Name D.ClusterName
		}
	}

	err = json.Unmarshal(body, &pcl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse json at %s", url)
	}

	cls := pcl.Clusters[account]

	clusters = make([]D.ClusterName, len(cls))
	for i, cl := range cls {
		clusters[i] = cl.Name
	}

	return clusters, nil
}

// GetRegionNames returns a list of regions that a cluster is deployed into
func (s Spinnaker) GetRegionNames(app string, account D.AccountName, cluster D.ClusterName) ([]D.RegionName, error) {
	url := s.clusterURL(app, string(account), string(cluster))
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	var cl struct {
		ServerGroups []struct{ Region D.RegionName }
	}

	err = json.Unmarshal(body, &cl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse json at %s", url)
	}

	set := make(map[D.RegionName]bool)
	for _, g := range cl.ServerGroups {
		set[g.Region] = true
	}

	result := make([]D.RegionName, 0, len(set))
	for region := range set {
		result = append(result, region)
	}

	return result, nil
}
