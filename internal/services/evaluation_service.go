package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/PRM710/Rankedterview-backend/internal/config"
	"github.com/PRM710/Rankedterview-backend/internal/models"
)

type EvaluationService struct {
	openaiClient *openai.Client
	config       *config.Config
}

func NewEvaluationService(cfg *config.Config) *EvaluationService {
	client := openai.NewClient(cfg.OpenAIKey)
	return &EvaluationService{
		openaiClient: client,
		config:       cfg,
	}
}

// EvaluateInterview evaluates an interview using AI
func (s *EvaluationService) EvaluateInterview(ctx context.Context, transcript string) (*models.Evaluation, error) {
	if transcript == "" {
		return nil, errors.New("transcript is empty")
	}

	// Create evaluation prompt
	prompt := s.buildEvaluationPrompt(transcript)

	// Call OpenAI API
	resp, err := s.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.config.OpenAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert interview evaluator. Analyze the interview transcript and provide detailed feedback.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   s.config.OpenAIMaxTokens,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	// Parse the AI response
	evaluation, err := s.parseEvaluation(resp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}

	// Add metadata
	evaluation.ProcessedAt = time.Now()
	evaluation.AIModel = s.config.OpenAIModel
	evaluation.TokensUsed = resp.Usage.TotalTokens

	return evaluation, nil
}

// buildEvaluationPrompt creates the prompt for interview evaluation
func (s *EvaluationService) buildEvaluationPrompt(transcript string) string {
	return fmt.Sprintf(`
Analyze this interview transcript and provide a detailed evaluation.

TRANSCRIPT:
%s

Please evaluate the interview on the following criteria (score 0-100 for each):
1. Communication: Clarity, articulation, and effective expression
2. Technical: Accuracy and depth of technical knowledge
3. Confidence: Self-assurance and composure
4. Structure: Logical flow and organization of responses

Also provide:
- 3-5 key strengths
- 3-5 areas for improvement
- Overall summary (2-3 sentences)
- 2-3 timestamped highlights (good moments and areas to improve)

Format your response as JSON with this structure:
{
  "scores": {
    "communication": 0-100,
    "technical": 0-100,
    "confidence": 0-100,
    "structure": 0-100,
    "overall": 0-100
  },
  "feedback": {
    "strengths": ["strength 1", "strength 2", ...],
    "improvements": ["improvement 1", "improvement 2", ...],
    "summary": "overall summary",
    "highlights": [
      {"timestamp": 120.5, "type": "good", "comment": "excellent explanation"},
      {"timestamp": 305.2, "type": "improve", "comment": "could be clearer"}
    ]
  }
}
`, transcript)
}

// parseEvaluation parses the AI response into an Evaluation model
func (s *EvaluationService) parseEvaluation(aiResponse string) (*models.Evaluation, error) {
	// Try to extract JSON from response (AI might add explanation text)
	start := -1
	end := -1
	
	for i, char := range aiResponse {
		if char == '{' && start == -1 {
			start = i
		}
		if char == '}' {
			end = i + 1
		}
	}

	if start == -1 || end == -1 {
		return nil, errors.New("could not find JSON in AI response")
	}

	jsonStr := aiResponse[start:end]

	// Parse JSON response
	var result struct {
		Scores   models.Scores   `json:"scores"`
		Feedback models.Feedback `json:"feedback"`
	}

	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Calculate overall score if not provided
	if result.Scores.Overall == 0 {
		result.Scores.Overall = (result.Scores.Communication +
			result.Scores.Technical +
			result.Scores.Confidence +
			result.Scores.Structure) / 4.0
	}

	evaluation := &models.Evaluation{
		Scores:   result.Scores,
		Feedback: result.Feedback,
	}

	return evaluation, nil
}

// GenerateQuickFeedback generates quick feedback without full evaluation
func (s *EvaluationService) GenerateQuickFeedback(ctx context.Context, transcript string) (string, error) {
	resp, err := s.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.config.OpenAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an interview coach. Provide brief, actionable feedback.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Provide 3 quick tips to improve based on this interview:\n\n%s", transcript),
				},
			},
			MaxTokens:   500,
			Temperature: 0.8,
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
