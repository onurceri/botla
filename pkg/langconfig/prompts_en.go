package langconfig

const (
	EN_TopicExtractionSystemPrompt = "You are a helpful assistant."
	EN_TopicExtractionUserPrompt   = `The text above is a knowledge source for a chatbot. Write a single sentence summarizing what capabilities or information this text provides to the chatbot.
Only rely on the information present in the text. If the text is meaningless or contains no information, state that.
Example: "Provides information about the company's history and vision."
Summary:`
)
