package rag

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onurceri/botla-app/internal/models"
)

// fakeToolsLLM answers every completion with a fixed reply.
type fakeToolsLLM struct {
	reply     *string
	noChoices bool
	err       error
	tokens    int
	gotModel  string
	gotMsgs   []ChatMessage
}

func (f *fakeToolsLLM) CreateCompletion(context.Context, models.CompletionParams) (*models.CompletionResult, error) {
	return nil, errors.New("not used")
}
func (f *fakeToolsLLM) GetModelInfo() models.ModelInfo { return models.ModelInfo{} }
func (f *fakeToolsLLM) CreateCompletionWithTools(_ context.Context, msgs []ChatMessage, _ []Tool, model string, _ float32, _ int) (*ChatResponseWithTools, error) {
	f.gotModel, f.gotMsgs = model, msgs
	if f.err != nil {
		return nil, f.err
	}
	r := &ChatResponseWithTools{}
	r.Usage.TotalTokens = f.tokens
	if !f.noChoices {
		r.Choices = append(r.Choices, struct {
			Message      ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{Message: ChatMessage{Role: "assistant", Content: f.reply}})
	}
	return r, nil
}

func reply(s string) *fakeToolsLLM { return &fakeToolsLLM{reply: &s, tokens: 42} }

func TestDecomposeQuery(t *testing.T) {
	ctx := context.Background()
	q := "Who founded Halvane, and when did the Beacon S2 launch?"

	llm := reply(`{"queries": ["Who founded Halvane?", "When did the Beacon S2 launch?"]}`)
	got, tokens, err := DecomposeQuery(ctx, llm, "gpt-4o-mini", q)
	require.NoError(t, err)
	assert.Equal(t, []string{"Who founded Halvane?", "When did the Beacon S2 launch?"}, got)
	assert.Equal(t, 42, tokens)
	assert.Equal(t, "gpt-4o-mini", llm.gotModel)
	require.Len(t, llm.gotMsgs, 2)
	assert.Equal(t, "system", llm.gotMsgs[0].Role)
	assert.Contains(t, *llm.gotMsgs[1].Content, q, "the question is sent as data in the user turn")

	got, _, err = DecomposeQuery(ctx, reply("```json\n{\"queries\": [\"a?\", \"A?\", \" \", \"b?\", \"c?\", \"d?\"]}\n```"), "m", q)
	require.NoError(t, err)
	assert.Equal(t, []string{"a?", "b?", "c?"}, got, "fences stripped, blanks and duplicates dropped, capped at 3")
}

func TestDecomposeQuery_SinglePartKeepsQuestionVerbatim(t *testing.T) {
	ctx := context.Background()
	q := "Who is the CTO?"
	for _, r := range []string{`{"queries": ["Who is the CTO?"]}`, `{"queries": ["Who is the chief technology officer of Halvane?"]}`, `{"queries": []}`} {
		got, tokens, err := DecomposeQuery(ctx, reply(r), "m", q)
		require.NoError(t, err, r)
		assert.Equal(t, []string{q}, got, r)
		assert.Equal(t, 42, tokens)
	}
}

func TestDecomposeQuery_FailuresFallBackToQuestion(t *testing.T) {
	ctx := context.Background()
	q := "What is X, and what is Y?"

	got, tokens, err := DecomposeQuery(ctx, nil, "m", q)
	assert.Error(t, err)
	assert.Equal(t, []string{q}, got)
	assert.Zero(t, tokens)

	got, _, err = DecomposeQuery(ctx, &fakeToolsLLM{err: errors.New("boom")}, "m", q)
	assert.ErrorContains(t, err, "boom")
	assert.Equal(t, []string{q}, got)

	got, tokens, err = DecomposeQuery(ctx, &fakeToolsLLM{noChoices: true, tokens: 5}, "m", q)
	assert.Error(t, err)
	assert.Equal(t, []string{q}, got)
	assert.Equal(t, 5, tokens)

	got, _, err = DecomposeQuery(ctx, reply("Sure! Here are the queries: X and Y"), "m", q)
	assert.ErrorContains(t, err, "parse reply")
	assert.Equal(t, []string{q}, got)

	llm := reply(`{"queries": ["a", "b"]}`)
	got, _, err = DecomposeQuery(ctx, llm, "m", "   ")
	require.NoError(t, err)
	assert.Equal(t, []string{"   "}, got)
	assert.Nil(t, llm.gotMsgs, "blank questions are not sent")
}

func TestParseSubQueries_TruncatesByRune(t *testing.T) {
	long := strings.Repeat("ğ", maxSubQueryLen+50)
	got, err := parseSubQueries(`{"queries": ["` + long + `", "kısa"]}`)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, maxSubQueryLen, len([]rune(got[0])))
}
