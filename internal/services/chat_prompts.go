package services

import "fmt"

// =============================================================================
// LLM PROMPTS - ALWAYS IN ENGLISH
// Language-specific output is controlled via BuildLanguageDirective
// =============================================================================

// BaseSystemPrompt is the core prompt for chat completions.
// This prompt establishes the AI behavior and core rules.
const BaseSystemPrompt = `You are an AI assistant.

CORE RULES:
1. ONLY answer based on the provided source documents.
2. For topics not in your sources, say "I don't have information on this topic."
3. If user selects from a numbered list (e.g., "1", "2", "6"), understand from conversation context and respond appropriately.
4. Always consider previous conversation messages.`

// SmartFallbackPromptTemplate is used when no RAG context is available.
// %s placeholder is for capability text (can be empty).
const SmartFallbackPromptTemplate = `You are a customer support assistant. The user asked a question but you don't have information on this topic.

IMPORTANT RULES:
1. NEVER provide made-up or speculative information
2. Politely indicate you don't have information on this topic
3. If provided, mention what topics you CAN help with:
%s
4. When listing capabilities, provide a GENERAL SUMMARY, do not list items one by one.
5. Keep your response short and polite.`

// CapabilityInstructionEN is appended when capability summaries exist.
const CapabilityInstructionEN = `

### Available Resources and Capabilities:
You have access to the following resources. When explaining your capabilities, DO NOT list them one by one.
Instead, provide a GENERAL SUMMARY of the information these resources provide.
`

// RAGContextIntroEN is prepended to RAG context in user messages.
const RAGContextIntroEN = "The following documents are relevant to the query:\n\n"

// TopicExtractionSystemPromptEN for capability extraction from content.
const TopicExtractionSystemPromptEN = "You are a helpful assistant."

// TopicExtractionUserPromptTemplateEN for capability extraction.
// %s placeholder is for the target language name.
const TopicExtractionUserPromptTemplateEN = `The text above is a knowledge source for a chatbot. Write a single sentence summarizing what capabilities or information this text provides to the chatbot.
Only rely on the information present in the text. If the text is meaningless or contains no information, state that.
Example: "Provides information about the company's history and vision."

IMPORTANT: Write the summary in %s.
Summary:`

// TopicExtractionJSONPromptTemplateEN for structured metadata extraction.
// %s placeholder is for the target language name.
const TopicExtractionJSONPromptTemplateEN = `The text above is a knowledge source for a chatbot. Write a single sentence summarizing what capabilities or information this text provides to the chatbot.
Only rely on the information present in the text. If the text is meaningless or contains no information, state that.
Example: "Provides information about the company's history and vision."

Respond ONLY in JSON format:
{
  "capability_summary": <short sentence>,
  "suggested_questions": [<3-6 short and varied questions>]
}
No extra explanation or text. Write the summary and questions in %s.`

// CapabilityIntroEN for listing bot capabilities to users.
const CapabilityIntroEN = "I can help you with the following topics:"

// BuildLanguageDirective creates the language enforcement instruction.
// langName should be the full language name (e.g., "Turkish", "English").
func BuildLanguageDirective(langName string) string {
	return fmt.Sprintf("\n\nLANGUAGE REQUIREMENT: You MUST respond ONLY in %s. Never switch to another language.", langName)
}

// BuildSystemPrompt creates a complete system prompt with base rules,
// custom instructions, capabilities, and language enforcement.
func BuildSystemPrompt(customInstruction string, capabilities string, langName string) string {
	base := BaseSystemPrompt

	// Add custom instructions if provided
	if customInstruction != "" {
		base += "\n\n### Additional Instructions:\n" + customInstruction
	}

	// Add capabilities summary
	if capabilities != "" {
		base += CapabilityInstructionEN + capabilities
	}

	// Add language directive
	base += BuildLanguageDirective(langName)

	return base
}

// BuildSmartFallbackPrompt creates a prompt for the smart fallback mode.
func BuildSmartFallbackPrompt(capabilities string, langName string) string {
	capabilityText := ""
	if capabilities != "" {
		capabilityText = CapabilityIntroEN + "\n" + capabilities
	}
	prompt := fmt.Sprintf(SmartFallbackPromptTemplate, capabilityText)
	prompt += BuildLanguageDirective(langName)
	return prompt
}
