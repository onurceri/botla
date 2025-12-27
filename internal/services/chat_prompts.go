package services

import (
	"fmt"

	"github.com/onurceri/botla-co/internal/models"
)

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
func BuildSystemPrompt(botName string, customInstruction string, capabilities string, langName string, topicRestrictions *models.TopicConfig) string {
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

	// Add Topic Restrictions
	if topicRestrictions != nil {
		hasAllowed := len(topicRestrictions.AllowedTopics) > 0
		hasBlocked := len(topicRestrictions.BlockedTopics) > 0

		if hasAllowed || hasBlocked {
			base += "\n\n### STRICT TOPIC RESTRICTIONS / GUARDRAILS:\n"

			if hasAllowed {
				base += "You are STRICTLY LIMITED to discussing ONLY the following topics:\n"
				for _, topic := range topicRestrictions.AllowedTopics {
					base += fmt.Sprintf("- %s\n", topic)
				}
				// Explicitly mention greeting exception
				base += "- (You may still answer casual greetings like 'Hello' or 'How are you')\n"

				msg := "I'm sorry, I can only help with specific topics."
				if topicRestrictions.BlockedMessage != "" {
					msg = topicRestrictions.BlockedMessage
				}
				base += fmt.Sprintf("For ANY other topic (including asking for general information, recipes, code, etc. not listed above), you MUST refuse to answer and say exactly: \"%s\"\n", msg)
			}

			if hasBlocked {
				base += "You are STRICTLY FORBIDDEN from discussing the following topics:\n"
				for _, topic := range topicRestrictions.BlockedTopics {
					base += fmt.Sprintf("- %s\n", topic)
				}
			}
		}
	}

	// Add language directive
	base += BuildLanguageDirective(langName)

	return base
}

// RestrictedFallbackPromptTemplate is used when no RAG context is available.
// It allows natural conversation (greetings, identity questions) while
// preventing the bot from answering factual questions with general LLM knowledge.
// First %s = bot name, Second %s = capabilities list (may be empty)
const RestrictedFallbackPromptTemplate = `You are a friendly AI assistant named "%s".

## CORE BEHAVIOR

**ALWAYS ALLOWED - Respond warmly and naturally:**
- Greetings: "Merhaba!", "Selam!", "Hello!", "Nasılsın?", "Naber?"
- Identity questions: "Sen kimsin?", "Adın ne?", "What's your name?"
- Capability questions: "Ne yapabilirsin?", "How can you help me?"
- Basic small talk: "Teşekkürler", "İyi günler", "Güle güle"

**Example greeting response:** "Merhaba! Ben %s. Size nasıl yardımcı olabilirim?"

**When asked about capabilities, mention:**
%s

## STRICT RESTRICTIONS

**NEVER do these - even if the user insists:**
- Answer factual questions (products, prices, features, policies, comparisons)
- Provide information about companies, people, places, or events
- Make up, guess, or infer any specific information
- Write creative content (blogs, essays, code, poems)
- Give advice requiring expertise (medical, legal, financial)

**For any factual question, respond with something like:**
"Henüz bu konuda bilgi kaynaklarım yüklenmedi. Yakında size daha iyi yardımcı olabileceğim!"
(or in English: "My knowledge sources haven't been set up for this topic yet. I'll be able to help you better soon!")

## RESPONSE STYLE
- Be warm, friendly, and concise
- Keep responses to 1-2 sentences maximum
- Always respond in the same language the user is using`

// BuildRestrictedFallbackPrompt creates a highly restrictive prompt for fallback mode.
// This is used when no RAG context is found to prevent the bot from answering
// general questions using LLM's base knowledge.
func BuildRestrictedFallbackPrompt(botName string, capabilities string, langName string) string {
	capabilityText := "(No specific topics configured yet)"
	if capabilities != "" {
		capabilityText = capabilities
	}
	// Template uses botName twice: once in intro, once in example response
	prompt := fmt.Sprintf(RestrictedFallbackPromptTemplate, botName, botName, capabilityText)
	prompt += BuildLanguageDirective(langName)
	return prompt
}
