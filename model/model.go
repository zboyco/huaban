package model

type PageJson struct {
	Board *Board `json:"board"`
}

type Board struct {
	BoardID  int    `json:"board_id"`
	Title    string `json:"title"`
	Pins     []*Pin `json:"pins"`
	PinCount int    `json:"pin_count"`
}

type Pin struct {
	PinID int `json:"pin_id"`
	File  struct {
		Bucket string `json:"bucket"`
		Type   string `json:"type"`
		Key    string `json:"key"`
	} `json:"file"`
	Trusted bool `json:"trusted"`
}
