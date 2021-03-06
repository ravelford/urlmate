package main

import (
	"errors"
	"log"

	riak "github.com/basho/riak-go-client"
)

// RiakClient structure
type RiakClient struct {
	IP             string
	ShortURLBucket string
}

// CreateURL riak api
func (rc *RiakClient) CreateURL(key string, value []byte) (string, error) {
	data, err := rc.createObject(rc.ShortURLBucket, key, value)
	if err != nil {
		return "", err
	}
	return data.Key, nil
}

// GetURL riak api
func (rc *RiakClient) GetURL(key string) ([]byte, error) {
	data, err := rc.getObject(rc.ShortURLBucket, key)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

func (rc *RiakClient) client() *riak.Client {
	var err error

	o := &riak.NewClientOptions{
		RemoteAddresses: []string{rc.IP},
	}

	var c *riak.Client
	c, err = riak.NewClient(o)
	if err != nil {
		log.Println(err)
		return nil
	}
	return c
}

func (rc *RiakClient) createObject(bucket string, key string, value []byte) (*riak.Object, error) {
	c := rc.client()
	if c == nil {
		return nil, errors.New("no client")
	}
	defer func() {
		if err := c.Stop(); err != nil {
			log.Println(err)
		}
	}()

	object := &riak.Object{
		Bucket:      bucket,
		Key:         key,
		ContentType: "application/json",
		Charset:     "utf8",
		Value:       value,
	}
	cmd, err := riak.NewStoreValueCommandBuilder().
		WithContent(object).
		WithReturnBody(true).
		Build()
	if err != nil {
		return nil, err
	}
	if err = c.Execute(cmd); err != nil {
		return nil, err
	}
	scmd := cmd.(*riak.StoreValueCommand)
	if len(scmd.Response.Values) > 1 {
		return nil, errors.New("unexpected siblings in response")
	}
	return scmd.Response.Values[0], nil
}

func (rc *RiakClient) getObject(bucket string, key string) (*riak.Object, error) {
	c := rc.client()
	if c == nil {
		return nil, errors.New("no client")
	}
	defer func() {
		if err := c.Stop(); err != nil {
			log.Println(err)
		}
	}()

	cmd, err := riak.NewFetchValueCommandBuilder().
		WithBucket(bucket).
		WithKey(key).
		Build()
	if err != nil {
		return nil, err
	}
	if err := c.Execute(cmd); err != nil {
		return nil, err
	}

	fcmd := cmd.(*riak.FetchValueCommand)
	if len(fcmd.Response.Values) == 0 {
		return nil, errors.New("object not found")
	}

	if len(fcmd.Response.Values) > 1 {
		return nil, errors.New("unexpected siblings in response")
	}
	return fcmd.Response.Values[0], nil
}
