// Package ytsearch provides functionality for searching YouTube.
package ytsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bitly/go-simplejson"
)

const apiURL = "https://www.youtube.com/youtubei/v1"
const apiKey = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

func Search(query string) ([]Video, error) {
	type Client struct {
		Name    string `json:"clientName"`
		Version string `json:"clientVersion"`
	}
	type Context struct {
		Client Client `json:"client"`
	}
	data := struct {
		Context `json:"context"`
		Query   string `json:"query"`
	}{Context{Client{"WEB", "2.20201021.03.00"}}, query}
	var out json.RawMessage
	err := doAPIJSON("POST", "search",
		data, &out)
	if err != nil {
		return nil, err
	}
	vjson, err := simplejson.NewJson(out)
	if err != nil {
		return nil, err
	}
	vjson = vjson.GetPath("contents", "twoColumnSearchResultsRenderer", "primaryContents",
		"sectionListRenderer", "contents").GetIndex(0).GetPath("itemSectionRenderer", "contents")
	videos := make([]Video, len(vjson.MustArray()))
	videoslen := 0
	for i := range videos {
		jvideo := vjson.GetIndex(i).Get("videoRenderer")
		if jvideo.Interface() == nil {
			continue
		}
		videos[i].Title = jvideo.GetPath("title", "runs").GetIndex(0).Get("text").MustString()
		videos[i].ID = jvideo.Get("videoId").MustString()
		videoslen++
	}
	return videos[:videoslen], nil
}

type Video struct {
	Title string
	ID    string
}

func doAPIJSON(method, endpoint string, v, out interface{}) error {
	return doJSON(method,
		apiURL+"/"+endpoint+"?key="+apiKey, v, out)
}

func doJSON(method, url string, v, out interface{}) error {
	var body io.Reader
	if v != nil {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if 200 < resp.StatusCode || resp.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}
