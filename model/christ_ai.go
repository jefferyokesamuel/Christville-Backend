package model

type VerseRequest struct {
    Prompt string `json:"prompt" binding:"required"`
}

type Verse struct {
    Verse string `json:"verse"`
    Text  string `json:"text"`
}
