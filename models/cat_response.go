package models

type CatResponse struct {
	URLs  []string `json:"urls"`
	Count int      `json:"count"`
	Batch int      `json:"batch"`
}
