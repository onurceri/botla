package services

import "fmt"

// =============================================================================
// LLM PROMPTS - ALWAYS IN ENGLISH
// Language-specific output is controlled via BuildLanguageDirective
// =============================================================================

// BaseSystemPrompt is the core prompt for chat completions.
// This prompt establishes the AI behavior and core rules.
// %s placeholder is for bot name.
const BaseSystemPrompt = `You are an AI assistant named "%s".

## YOUR IDENTITY
- Your name is "%s"
- You are an AI assistant that helps users find information

## HOW TO RESPOND

**Greetings & Casual Conversation:**
Respond naturally. Share your name if asked. Be friendly and brief.

**Factual Questions (products, services, policies, prices, technical details, etc.):**
- Answer based on the provided source documents
- If information is not in sources, clearly state: "I don't have information on this specific topic"
- When your answer is based on limited or partial information, naturally express uncertainty (e.g., "Based on the available information...", "From what I can see...")
- NEVER invent facts, prices, or specific details

**Capability Questions ("What can you do?"):**
Give a brief, general summary of your purpose. Don't list every capability.

## CONTEXT HANDLING
- When source documents are provided, use them to answer factual questions
- When no relevant sources are found, be honest about what you don't know
- Understand numbered selections (1, 2, 6) from conversation context
- Always consider previous conversation messages`

// SmartFallbackPromptTemplate is used when no RAG context is available.
// First %s is for bot name, second %s is for capability text (can be empty).
const SmartFallbackPromptTemplate = `You are a customer support assistant named "%s". The user asked a question but you don't have information on this topic.

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
// botName is used in the identity section of the prompt.
func BuildSystemPrompt(botName string, customInstruction string, capabilities string, langName string) string {
	// Bot name appears twice in the template (intro and identity section)
	base := fmt.Sprintf(BaseSystemPrompt, botName, botName)

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
// botName is included for consistent identity.
func BuildSmartFallbackPrompt(botName string, capabilities string, langName string) string {
	capabilityText := ""
	if capabilities != "" {
		capabilityText = CapabilityIntroEN + "\n" + capabilities
	}
	prompt := fmt.Sprintf(SmartFallbackPromptTemplate, botName, capabilityText)
	prompt += BuildLanguageDirective(langName)
	return prompt
}

// RestrictedFallbackPromptTemplate is a stricter version used when no RAG context is available.
// This prevents the bot from using general LLM knowledge to answer factual questions.
const RestrictedFallbackPromptTemplate = `You are a customer support assistant named "%s".

STRICT RULES - FOLLOW EXACTLY:

ALLOWED:
- Respond naturally to greetings (e.g., "Hello!", "How are you?")
- Say your name if asked
- If asked what you can help with, briefly mention these topics:
%s

STRICTLY FORBIDDEN:
- Answering ANY factual question (products, prices, features, comparisons, recommendations)
- Providing information about companies, people, events, or general knowledge
- Making up or guessing any information
- Giving detailed explanations about any topic

For any factual question, politely say you don't have information on that topic.
Keep responses under 2 sentences.`

// BuildRestrictedFallbackPrompt creates a highly restrictive prompt for fallback mode.
// This is used when no RAG context is found to prevent the bot from answering
// general questions using LLM's base knowledge.
func BuildRestrictedFallbackPrompt(botName string, capabilities string, langName string) string {
	capabilityText := ""
	if capabilities != "" {
		capabilityText = capabilities
	}
	prompt := fmt.Sprintf(RestrictedFallbackPromptTemplate, botName, capabilityText)
	prompt += BuildLanguageDirective(langName)
	return prompt
}


