package models

type CatProfile struct {
    ID          int      `json:"id"`
    Img         string   `json:"img"`
    Name        string   `json:"name"`
    Age         int      `json:"age"`
    Breed       string   `json:"breed"`
    Personality string   `json:"personality"`
    Hobbies     []string `json:"hobbies"`
    Bio         string   `json:"bio"`
}