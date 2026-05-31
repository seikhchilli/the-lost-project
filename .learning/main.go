package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"
)

type Target struct {
	Credits Credits `json:"credits"`
}

type Credits struct {
	Cast []Cast `json:"cast"`
}

type Cast struct {
	Name       string  `json:"name"`
	Popularity float64 `json:"popularity"`
}

func main() {
	client := http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/movie/27205?api_key=6237a1b914c72a7266ad8783ee84571b&append_to_response=external_ids,credits", nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	var target Target

	json.NewDecoder(resp.Body).Decode(&target)

	casts := target.Credits.Cast

	slices.SortFunc(casts, func(a, b Cast) int { return cmp.Compare(b.Popularity, a.Popularity) })
	fmt.Println(casts)
}
