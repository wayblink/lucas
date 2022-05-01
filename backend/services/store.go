package services

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewWithOutTLS(endpoints string) (*Store, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Errorf("Failed to create etcd client,because of %v", err.Error())
	}
	return &Store{
		Client: client,
	}, err
}

func New(endpoints string, ca string, key string, cert string) (*Store, error) {
	if endpoints == "" {
		log.Errorf("Failed to get etcd endpoints from command line args")
		return nil, errors.New("Endpoint need provide.")
	}

	if ca != "" && key != "" && cert != "" {
		tlsConf, err := createTlsConf(ca, key, cert)
		if err != nil {
			log.Panicf("Failed to crate tls config, because of %s", err.Error())
		}
		client, err := clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(endpoints, ","),
			TLS:         tlsConf,
			DialTimeout: 5 * time.Second,
		})
		return &Store{
			Client: client,
		}, nil
	} else {
		return NewWithOutTLS(endpoints)
	}
}

func createTlsConf(ca, key, cert string) (*tls.Config, error) {
	cfgtls := &transport.TLSInfo{}
	//cfgtls.CAFile = ca
	cfgtls.KeyFile = key
	cfgtls.CertFile = cert
	clientTLS, err := cfgtls.ClientConfig()
	if err != nil {
		return nil, err
	}
	//add default InsecureSkipVerify
	clientTLS.InsecureSkipVerify = true
	return clientTLS, nil
}

type Store struct {
	Client *clientv3.Client
}

func (store *Store) List(key string) []*KeyValue {
	opts := []clientv3.OpOption{}
	opts = append(opts, clientv3.WithPrefix(), clientv3.WithSerializable())
	keyValues := make([]*KeyValue, 0)
	resp, err := store.Client.Get(context.Background(), key, opts...)
	if err != nil {
		log.Errorf("Failed to list keys from etcd, because of %v", err.Error())
		return nil
	}
	for _, kv := range resp.Kvs {
		key_str := string(kv.Key)
		value_str := string(kv.Value)
		keyValue := &KeyValue{
			Key:            key_str,
			CreateRevision: kv.CreateRevision,
			ModRevision:    kv.ModRevision,
			Version:        kv.Version,
			Lease:          kv.Lease,
			Value:          value_str,
		}
		keyValues = append(keyValues, keyValue)
	}
	return keyValues
}

func (store *Store) ListByPath(key string) *Tree {
	tree := &Tree{}
	keyValues := store.List(key)
	for _, kv := range keyValues {
		tree.AddNode(kv)
	}
	return tree
}

func (store *Store) Put(key string, value string) error {
	opts := []clientv3.OpOption{}
	_, err := store.Client.Put(context.Background(), key, value, opts...)
	if err != nil {
		log.Errorf("Failed to put key to etcd, because of %v", err.Error())
		return err
	}
	return err
}

func (store *Store) Delete(key string) error {
	opts := []clientv3.OpOption{}
	_, err := store.Client.Delete(context.Background(), key, opts...)
	if err != nil {
		log.Errorf("Failed to delete key to etcd, because of %v", err.Error())
		return err
	}
	return err
}

/**
mock a etcd pb result
*/

type KeyValue struct {
	// key is the key in bytes. An empty key is not allowed.
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// create_revision is the revision of last creation on this key.
	CreateRevision int64 `protobuf:"varint,2,opt,name=create_revision,json=createRevision,proto3" json:"create_revision,omitempty"`
	// mod_revision is the revision of last modification on this key.
	ModRevision int64 `protobuf:"varint,3,opt,name=mod_revision,json=modRevision,proto3" json:"mod_revision,omitempty"`
	// version is the version of the key. A deletion resets
	// the version to zero and any modification of the key
	// increases its version.
	Version int64 `protobuf:"varint,4,opt,name=version,proto3" json:"version,omitempty"`
	// value is the value held by the key, in bytes.
	Value string `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	// lease is the ID of the lease that attached to key.
	// When the attached lease expires, the key will be deleted.
	// If lease is 0, then no lease is attached to the key.
	Lease int64 `protobuf:"varint,6,opt,name=lease,proto3" json:"lease,omitempty"`
}

type Tree struct {
	Nodes map[string]*Node
}

type Node struct {
	RootPath string
	Nodes    map[string]*Node
	KV       *KeyValue
}

func (n *Node) AddLeaf(kv *KeyValue) {
	sn := &Node{}
	if n.Nodes == nil {
		n.Nodes = make(map[string]*Node)
	}
	subPath := kv.Key[len(n.RootPath):]
	if strings.Contains(subPath, "/") {
		index := strings.Index(subPath, "/")
		sn.RootPath = n.RootPath + subPath[:index+1]
		if n.Nodes[sn.RootPath] != nil {
			sn = n.Nodes[sn.RootPath]
		} else {
			n.Nodes[sn.RootPath] = sn
		}
		sn.AddLeaf(kv)
	} else {
		sn.RootPath = kv.Key
		sn.KV = kv
		n.Nodes[sn.RootPath] = sn
	}
}

func (t *Tree) AddNode(kv *KeyValue) {
	key := kv.Key
	if t.Nodes == nil {
		t.Nodes = make(map[string]*Node)
	}
	if strings.Contains(key, "/") {
		node := t.Nodes["/"]
		if node == nil {
			node = &Node{
				RootPath: "/",
			}
			t.Nodes["/"] = node
		}

		node.AddLeaf(kv)

	} else {
		t.Nodes[key] = &Node{
			RootPath: key,
			KV:       kv,
		}
	}
}
