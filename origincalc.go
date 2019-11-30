package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func GetOriginReqCount(in io.Reader, filterByHost string, startTs int, endTs int) (int, error) {
	total := 0
	counts := make(map[string]map[string]int, 0)
	dec := json.NewDecoder(in)
	err := dec.Decode(&counts)
	if err != nil {
		return -1, err
	}
	for ts, rec := range counts {
		tsInt, err := strconv.Atoi(ts)
		if err != nil {
			continue
		}
		if tsInt < startTs || tsInt > endTs {
			continue
		}

		for host, count := range rec {
			if host == filterByHost {
				total += count
				break
			}
		}

	}
	return total, nil
}

func GetOriginReqCountURL(srcURL string, filterByHost string, startTs int, endTs int) (int, error) {
	resp, err := http.Get(srcURL)

	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	return GetOriginReqCount(resp.Body, filterByHost, startTs, endTs)

}
