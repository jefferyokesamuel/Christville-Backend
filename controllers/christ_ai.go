package controllers

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "christville/model"
    "strings"

    "github.com/gin-gonic/gin"
)

// GetVerse handles the request to fetch Bible verses based on a user prompt.
func GetVerse(c *gin.Context) {
    var request model.VerseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Get topic or verse reference from AI
    topicOrVerse, err := getTopicOrVerseFromAI(request.Prompt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "AI processing failed"})
        return
    }

    // Fetch the actual verse(s) using the Bible API
    verses, err := getVersesFromBibleAPI(topicOrVerse)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch verses"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "prompt": request.Prompt,
        "result": verses,
    })
}

// getTopicOrVerseFromAI queries the AI to suggest a Bible topic or verse.
func getTopicOrVerseFromAI(prompt string) (string, error) {
    url := "https://api.openai.com/v1/completions"
    payload := `{
        "model": "GPT-3.5 Turbo",
        "prompt": "Suggest a Bible topic or verse for this prompt: ` + prompt + `",
        "max_tokens": 20
    }`

    client := &http.Client{}
    req, err := http.NewRequest("POST", url, strings.NewReader(payload))
    if err != nil {
        return "", err
    }

    req.Header.Set("Authorization", "Bearer YOUR_OPENAI_API_KEY")
    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return "", err
    }

    choices := result["choices"].([]interface{})
    if len(choices) == 0 {
        return "", errors.New("no response from AI")
    }

    output := choices[0].(map[string]interface{})["text"].(string)
    return strings.TrimSpace(output), nil
}

// getVersesFromBibleAPI queries the Bible API to fetch verses based on a query.
func getVersesFromBibleAPI(query string) ([]model.Verse, error) {
    apiURL := fmt.Sprintf("https://api.openbible.info/topics/%s", url.QueryEscape(query))

    resp, err := http.Get(apiURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var verses []model.Verse
    err = json.NewDecoder(resp.Body).Decode(&verses)
    if err != nil {
        return nil, err
    }

    return verses, nil
}
